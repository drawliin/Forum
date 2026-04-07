package db

import (
	"forum/internal/models"
)

// FetchComments returns all comments for one post in time order.
func FetchComments(postID int) ([]models.Comment, error) {
	rows, err := Database.Query(
		`SELECT c.id, c.content, c.user_id, u.username, c.created_at,
        (SELECT COUNT(*) FROM comment_reactions cr WHERE cr.comment_id = c.id AND cr.value = 1) AS likes,
        (SELECT COUNT(*) FROM comment_reactions cr WHERE cr.comment_id = c.id AND cr.value = -1) AS dislikes
        FROM comments c
        JOIN users u ON u.id = c.user_id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(
			&c.ID,
			&c.Content,
			&c.UserID,
			&c.Author,
			&c.CreatedAt,
			&c.Likes,
			&c.Dislikes,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// PostIDByComment finds the parent post of a comment.
func PostIDByComment(commentID int) (int, error) {
	var postID int
	err := Database.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	return postID, err
}
