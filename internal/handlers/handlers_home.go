package handlers

import (
	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"
	"net/http"
	"strconv"
)

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
		util.ClientError(w, r, http.StatusUnauthorized, "Please log in to use this filter")
		return
	}

	categoryID := 0
	if value := r.URL.Query().Get("category"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
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

	posts, err := db.FetchPosts(user, categoryID, filter)
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
	}
	templates.Render(w, "home", data, 0)
}
