package db

import (
	"strings"

	"forum/internal/models"
)

func FetchPosts(user *models.User, categoryID int, filter string) ([]models.Post, error) {
	query := `SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = 1) AS likes,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = -1) AS dislikes
        FROM posts p
        JOIN users u ON u.id = p.user_id`

	where := []string{"1=1"}
	args := []any{}

	if categoryID > 0 {
		where = append(where, "EXISTS (SELECT 1 FROM post_categories pc WHERE pc.post_id = p.id AND pc.category_id = ?)")
		args = append(args, categoryID)
	}
	if filter == "mine" && user != nil {
		where = append(where, "p.user_id = ?")
		args = append(args, user.ID)
	}
	if filter == "liked" && user != nil {
		where = append(where, "EXISTS (SELECT 1 FROM post_reactions pr WHERE pr.post_id = p.id AND pr.user_id = ? AND pr.value = 1)")
		args = append(args, user.ID)
	}

	query = query + " WHERE " + strings.Join(where, " AND ") + " ORDER BY p.created_at DESC"

	rows, err := Database.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.Author,
			&post.CreatedAt,
			&post.Likes,
			&post.Dislikes,
		); err != nil {
			rows.Close()
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for i := range posts {
		categories, err := FetchPostCategories(posts[i].ID)
		if err != nil {
			return nil, err
		}
		posts[i].Categories = categories
	}

	return posts, nil
}

func FetchPostByID(postID int) (*models.Post, error) {
	var post models.Post
	err := Database.QueryRow(
		`SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = 1) AS likes,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = -1) AS dislikes
        FROM posts p
        JOIN users u ON u.id = p.user_id
        WHERE p.id = ?`,
		postID,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.Author,
		&post.CreatedAt,
		&post.Likes,
		&post.Dislikes,
	)
	if err != nil {
		return nil, err
	}

	categories, err := FetchPostCategories(post.ID)
	if err != nil {
		return nil, err
	}
	post.Categories = categories
	return &post, nil
}

func FetchPostReaction(postID int) (likes, dislikes int, err error) {
	// var post models.Post
	err = Database.QueryRow(
		`SELECT
		SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END ) AS likes,
        SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END ) AS dislikes
        FROM post_reactions
        WHERE post_id = ?`,
		postID,
	).Scan(
		&likes,
		&dislikes,
	)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}
