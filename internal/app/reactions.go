package app

import (
    "database/sql"
    "errors"
    "time"
)

func (app *App) togglePostReaction(userID, postID, value int) error {
    var existing int
    err := app.db.QueryRow(
        "SELECT value FROM post_reactions WHERE post_id = ? AND user_id = ?",
        postID,
        userID,
    ).Scan(&existing)
    if errors.Is(err, sql.ErrNoRows) {
        _, err = app.db.Exec(
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
        _, err = app.db.Exec(
            "DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?",
            postID,
            userID,
        )
        return err
    }

    _, err = app.db.Exec(
        "UPDATE post_reactions SET value = ?, created_at = ? WHERE post_id = ? AND user_id = ?",
        value,
        time.Now().Unix(),
        postID,
        userID,
    )
    return err
}

func (app *App) toggleCommentReaction(userID, commentID, value int) error {
    var existing int
    err := app.db.QueryRow(
        "SELECT value FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
        commentID,
        userID,
    ).Scan(&existing)
    if errors.Is(err, sql.ErrNoRows) {
        _, err = app.db.Exec(
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
        _, err = app.db.Exec(
            "DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
            commentID,
            userID,
        )
        return err
    }

    _, err = app.db.Exec(
        "UPDATE comment_reactions SET value = ?, created_at = ? WHERE comment_id = ? AND user_id = ?",
        value,
        time.Now().Unix(),
        commentID,
        userID,
    )
    return err
}
