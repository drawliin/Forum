package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"
)

// addComment saves a new comment under a post.
func addComment(w http.ResponseWriter, r *http.Request, postID int) {
	user, ok := util.RequireAuth(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		util.ClientError(w, r, http.StatusBadRequest, "Invalid form")
		return
	}
	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		post, err := db.FetchPostByID(postID)
		if err != nil {
			util.ServerError(w, r, "Failed to load post")
			return
		}
		// Empty comment error on same page
		templates.Render(w, "post_view", models.TemplateData{
			FormError: "Comment cannot be empty",
			User:      user,
			Post:      post,
		}, http.StatusBadRequest)
		return
	} else if len(content) > 1028 {
		util.ClientError(w, r, http.StatusBadRequest, "Comment too long")
		return
	}

	if _, err := db.Database.Exec(
		"INSERT INTO comments (post_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
		postID,
		user.ID,
		content,
		time.Now().Unix(),
	); err != nil {
		util.ServerError(w, r, "Failed to add comment")
		return
	}
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

// reactToComment handles like and dislike actions on comments.
func reactToComment(w http.ResponseWriter, r *http.Request, commentID int) {
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

	if err := toggleCommentReaction(user.ID, commentID, value); err != nil {
		util.ServerError(w, r, "Failed to react to comment")
		return
	}

	//Fetch comment's reactions
	likes, dislikes, err := db.FetchCommentReaction(commentID)
	if err != nil {
		util.ServerError(w, r, "Failed to load comment ractions")
		return
	}

	// Send Likes and Dislikes in Json resp
	WriteJson(w, likes, dislikes)
}

// toggleCommentReaction adds, removes, or swaps a user's reaction on a comment.
func toggleCommentReaction(userID, commentID, value int) error {
	var existing int
	err := db.Database.QueryRow(
		"SELECT value FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
		commentID,
		userID,
	).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = db.Database.Exec(
			"INSERT INTO comment_reactions (comment_id, user_id, value, created_at) VALUES (?, ?, ?, ?)",
			commentID,
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
			"DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
			commentID,
			userID,
		)
		return err
	}

	_, err = db.Database.Exec(
		"UPDATE comment_reactions SET value = ?, created_at = ? WHERE comment_id = ? AND user_id = ?",
		value,
		time.Now().Unix(),
		commentID,
		userID,
	)
	return err
}

// commentHandler parses the comment route and forwards to the right action.
func commentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/comment/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}
	commentID, err := strconv.Atoi(parts[0])
	if err != nil || commentID < 1 {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}
	if parts[1] != "react" {
		util.ClientError(w, r, http.StatusNotFound, "Page not found")
		return
	}
	reactToComment(w, r, commentID)
}
