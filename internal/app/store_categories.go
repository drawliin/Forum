package app

func (app *App) fetchCategories() ([]Category, error) {
    rows, err := app.db.Query("SELECT id, name FROM categories ORDER BY name")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []Category
    for rows.Next() {
        var c Category
        if err := rows.Scan(&c.ID, &c.Name); err != nil {
            return nil, err
        }
        categories = append(categories, c)
    }
    return categories, rows.Err()
}

func (app *App) fetchPostCategories(postID int) ([]Category, error) {
    rows, err := app.db.Query(
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

    var categories []Category
    for rows.Next() {
        var c Category
        if err := rows.Scan(&c.ID, &c.Name); err != nil {
            return nil, err
        }
        categories = append(categories, c)
    }
    return categories, rows.Err()
}

func (app *App) categoryIDSet() (map[int]struct{}, error) {
    rows, err := app.db.Query("SELECT id FROM categories")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    ids := make(map[int]struct{})
    for rows.Next() {
        var id int
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        ids[id] = struct{}{}
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return ids, nil
}
