package db

import (
	"forum/internal/models"
)

func FetchCategories() ([]models.Category, error) {
	rows, err := Database.Query("SELECT id, name FROM categories ORDER BY name")
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

func FetchPostCategories(postID int) ([]models.Category, error) {
	rows, err := Database.Query(
		`SELECT c.id, c.name
        FROM categories c
        JOIN post_categories pc ON pc.category_id = c.id
        WHERE pc.post_id = ?
        ORDER BY c.name`,
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
