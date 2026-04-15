package handlers

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"
)

// homeHandler loads the feed page with optional filters from the query string.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}
	if r.Method != http.MethodGet {
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	user, err := util.CurrentUser(w, r)
	if err != nil {
		util.ServerError(w, r, "Failed to load session")
		return
	}

	filter := r.URL.Query().Get("filter")
	if filter != "" && filter != "mine" && filter != "liked" {
		util.ClientError(w, r, http.StatusBadRequest, "Invalid filter")
		return
	}
	if filter != "" && user == nil {
		util.ClientError(w, r, http.StatusUnauthorized, "Please log in first")
		return
	}

	validIDs, err := db.CategoryIDSet()
	if err != nil {
		util.ServerError(w, r, "Failed to load categories")
		return
	}

	categoryID := 0
	if value := r.URL.Query().Get("category"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || !slices.Contains(validIDs, parsed) {
			util.ClientError(w, r, http.StatusBadRequest, "Invalid category")
			return
		}
		categoryID = parsed
	}

	categories, err := db.FetchCategories()
	if err != nil {
		util.ServerError(w, r, "Failed to load categories")
		return
	}

	page := 1
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 {
			util.ClientError(w, r, http.StatusBadRequest, "Invalid page")
			return
		}
	}

	offset := (page - 1) * db.PageSize
	posts, err := db.FetchPosts(user, categoryID, filter, offset)
	if err != nil {
		util.ServerError(w, r, "Failed to load posts")
		return
	}

	data := models.TemplateData{
		User:       user,
		Categories: categories,
		Posts:      posts,
		Filter:     filter,
		CategoryID: categoryID,
		Page:       page,
	}

	if len(data.Posts) > db.PageSize {
		data.HasNext = true
		data.Posts = data.Posts[:db.PageSize]
	}
	if page > 1 {
		data.HasPrev = true
		data.PrevPage = page - 1
	}
	if data.HasNext {
		data.NextPage = page + 1
	}

	params := url.Values{}
	if filter != "" {
		params.Set("filter", filter)
	}
	if categoryID > 0 {
		params.Set("category", strconv.Itoa(categoryID))
	}
	if encoded := params.Encode(); encoded != "" {
		data.PageQuery = "&" + encoded
	}
	templates.Render(w, "home", data, 0)
}
