package app

import "strings"

func (app *App) fetchPosts(user *User, categoryID int, filter string) ([]Post, error) {
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

    rows, err := app.db.Query(query, args...)
    if err != nil {
        return nil, err
    }

    var posts []Post
    for rows.Next() {
        var post Post
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
        categories, err := app.fetchPostCategories(posts[i].ID)
        if err != nil {
            return nil, err
        }
        posts[i].Categories = categories
    }

    return posts, nil
}

func (app *App) fetchPostByID(postID int) (*Post, error) {
    var post Post
    err := app.db.QueryRow(
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

    categories, err := app.fetchPostCategories(post.ID)
    if err != nil {
        return nil, err
    }
    post.Categories = categories
    return &post, nil
}
