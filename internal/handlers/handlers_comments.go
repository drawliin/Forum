package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/db"
	"forum/internal/util"
)

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
		util.ClientError(w, r, http.StatusBadRequest, "Comment cannot be empty")
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

	postID, err := db.PostIDByComment(commentID)
	if err != nil {
		util.ServerError(w, r, "Failed to reload post")
		return
	}
	http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

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
