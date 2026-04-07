package db

import (
	"forum/internal/models"
	"strings"
)

// FetchPosts builds the home feed with optional user and category filters.
func FetchPosts(user *models.User, categoryID int, filter string) ([]models.Post, error) {
	query := `SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = 1) AS likes,
        (SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = -1) AS dislikes
        FROM posts p
        JOIN users u ON u.id = p.user_id`

	where := []string{}
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

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := Database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	postIDs := make([]int, 0)
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
			return nil, err
		}
		posts = append(posts, post)
		postIDs = append(postIDs, post.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	postCategories, err := FetchCategoriesForPosts(postIDs)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].Categories = postCategories[posts[i].ID]
	}

	return posts, nil
}

// FetchPostByID loads one post together with its category list.
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
