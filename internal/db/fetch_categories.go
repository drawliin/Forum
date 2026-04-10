package db

import (
	"forum/internal/models"
	"strings"
)

// FetchCategories returns all categories for forms and filters.
func FetchCategories() ([]models.Category, error) {
	rows, err := Database.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// FetchPostCategories returns the categories attached to one post.
func FetchPostCategories(postID int) ([]models.Category, error) {
	rows, err := Database.Query(
		`SELECT c.id, c.name
        FROM categories c
        JOIN post_categories pc ON pc.category_id = c.id
        WHERE pc.post_id = ?`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// FetchCategoriesForPosts loads categories for many posts in one query.
func FetchCategoriesForPosts(postIDs []int) (map[int][]models.Category, error) {
	postCategories := make(map[int][]models.Category, len(postIDs))
	if len(postIDs) == 0 {
		return postCategories, nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]any, len(postIDs))
	for i, postID := range postIDs {
		placeholders[i] = "?"
		args[i] = postID
	}

	rows, err := Database.Query(
		`SELECT pc.post_id, c.id, c.name
        FROM post_categories pc
        JOIN categories c ON c.id = pc.category_id
        WHERE pc.post_id IN (`+strings.Join(placeholders, ", ")+`)
        ORDER BY pc.post_id, c.id`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID int
		var category models.Category
		if err := rows.Scan(&postID, &category.ID, &category.Name); err != nil {
			return nil, err
		}
		postCategories[postID] = append(postCategories[postID], category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, postID := range postIDs {
		if _, ok := postCategories[postID]; !ok {
			postCategories[postID] = []models.Category{}
		}
	}

	return postCategories, nil
}

// CategoryIDSet returns category ids so handlers can validate user input.
func CategoryIDSet() ([]int, error) {
	rows, err := Database.Query("SELECT id FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
