package db

import (
	"forum/internal/models"
)

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

func PostIDByComment(commentID int) (int, error) {
	var postID int
	err := Database.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	return postID, err
}

//Fetch comment gets the number of likes and dislikes of a comment based on its ID
//the query eturns 0 if a NULL value is found
func FetchCommentReaction(commentID int) (likes, dislikes int, err error) {
	err = Database.QueryRow(
		`SELECT
		COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END ),0) AS likes,
        COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END ),0) AS dislikes
        FROM comment_reactions
        WHERE comment_id = ?`,
		commentID,
	).Scan(
		&likes,
		&dislikes,
	)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}