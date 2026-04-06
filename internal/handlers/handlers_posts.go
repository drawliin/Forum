package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"
)

func postNewHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := util.RequireAuth(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		categories, err := db.FetchCategories()
		if err != nil {
			util.ServerError(w, r, "Failed to load categories")
			return
		}
		templates.Render(w, "post_new", models.TemplateData{User: user, Categories: categories}, 0)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			util.ClientError(w, r, http.StatusBadRequest, "Invalid form")
			return
		}
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		categoryValues := r.Form["categories"]

		if title == "" || content == "" {
			util.ClientError(w, r, http.StatusBadRequest, "Title and content are required")
			return
		}

		if len(title) > 65 {
			util.ClientError(w, r, http.StatusBadRequest, "Title too long")
			return
		}

		validIDs, err := db.CategoryIDSet()
		if err != nil {
			util.ServerError(w, r, "Failed to load categories")
			return
		}

		categoryIDs := make([]int, 0, len(categoryValues))
		seen := make(map[int]bool)

		for _, value := range categoryValues {
			id, err := strconv.Atoi(value)
			if err != nil || !slices.Contains(validIDs, id) {
				util.ClientError(w, r, http.StatusBadRequest, "Invalid category")
				return
			}
			if _, ok := seen[id]; ok {
				continue // ignore duplicated category
			}
			seen[id] = true
			categoryIDs = append(categoryIDs, id)
		}

		if len(categoryIDs) == 0 {
			util.ClientError(w, r, http.StatusBadRequest, "Select at least one valid category")
			return
		}

		tx, err := db.Database.BeginTx(context.Background(), nil)
		if err != nil {
			util.ServerError(w, r, "Failed to create post")
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec(
			"INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)",
			user.ID,
			title,
			content,
			time.Now().Unix(),
		)
		if err != nil {
			util.ServerError(w, r, "Failed to create post")
			return
		}
		postID64, err := res.LastInsertId()
		if err != nil {
			util.ServerError(w, r, "Failed to create post")
			return
		}

		for _, id := range categoryIDs {
			if _, err := tx.Exec(
				"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
				postID64,
				id,
			); err != nil {
				util.ServerError(w, r, "Failed to assign categories")
				return
			}
		}

		if err := tx.Commit(); err != nil {
			util.ServerError(w, r, "Failed to save post")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/post/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}

	postID, err := strconv.Atoi(parts[0])
	if err != nil || postID < 1 {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		if r.Method != http.MethodGet {
			util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		viewPost(w, r, postID)
	case "comment":
		if r.Method != http.MethodPost {
			util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		addComment(w, r, postID)
	case "react":
		if r.Method != http.MethodPost {
			util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		reactToPost(w, r, postID)
	default:
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
	}
}

func viewPost(w http.ResponseWriter, r *http.Request, postID int) {
	user, err := util.CurrentUser(w, r)
	if err != nil {
		util.ServerError(w, r, "Failed to load session")
		return
	}

	post, err := db.FetchPostByID(postID)
	if errors.Is(err, sql.ErrNoRows) {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}
	if err != nil {
		util.ServerError(w, r, "Failed to load post")
		return
	}

	comments, err := db.FetchComments(postID)
	if err != nil {
		util.ServerError(w, r, "Failed to load comments")
		return
	}

	categories, err := db.FetchCategories()
	if err != nil {
		util.ServerError(w, r, "Failed to load categories")
		return
	}

	data := models.TemplateData{
		User:       user,
		Categories: categories,
		Post:       post,
		Comments:   comments,
	}
	templates.Render(w, "post_view", data, 0)
}

func reactToPost(w http.ResponseWriter, r *http.Request, postID int) {
	user, ok := util.RequireAuth(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		util.ClientError(w, r, http.StatusBadRequest, "Invalid form")
		return
	}
	value, err := strconv.Atoi(r.FormValue("value"))
	if err != nil || (value != 1 && value != -1) {
		util.ClientError(w, r, http.StatusBadRequest, "Invalid reaction")
		return
	}

	if err := togglePostReaction(user.ID, postID, value); err != nil {
		util.ServerError(w, r, "Failed to react to post")
		return
	}

	referer := r.Referer()
	u, err := url.Parse(referer)
	if err != nil {
		util.ClientError(w, r, http.StatusBadRequest, "invalid url ")
	}

	path := u.Path
	if strings.HasPrefix(path, "/post/") {
		http.Redirect(w, r, path, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func togglePostReaction(userID, postID, value int) error {
	var existing int
	err := db.Database.QueryRow(
		"SELECT value FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID,
		userID,
	).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = db.Database.Exec(
			"INSERT INTO post_reactions (post_id, user_id, value, created_at) VALUES (?, ?, ?, ?)",
			postID,
			userID,
			value,
			time.Now().Unix(),
		)
		return err
	}
	if err != nil {
		return err
	}

	if existing == value {
		_, err = db.Database.Exec(
			"DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?",
			postID,
			userID,
		)
		return err
	}

	_, err = db.Database.Exec(
		"UPDATE post_reactions SET value = ?, created_at = ? WHERE post_id = ? AND user_id = ?",
		value,
		time.Now().Unix(),
		postID,
		userID,
	)
	return err
}
