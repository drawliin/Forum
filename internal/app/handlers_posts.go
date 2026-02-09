package app

import (
    "context"
    "database/sql"
    "errors"
    "net/http"
    "strconv"
    "strings"
    "time"
)

func (app *App) postNewHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := app.requireAuth(w, r)
    if !ok {
        return
    }

    switch r.Method {
    case http.MethodGet:
        categories, err := app.fetchCategories()
        if err != nil {
            app.serverError(w, r, "Failed to load categories")
            return
        }
        app.render(w, "post_new", TemplateData{User: user, Categories: categories}, 0)
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            app.clientError(w, r, http.StatusBadRequest, "Invalid form")
            return
        }
        title := strings.TrimSpace(r.FormValue("title"))
        content := strings.TrimSpace(r.FormValue("content"))
        categoryValues := r.Form["categories"]

        if title == "" || content == "" {
            app.clientError(w, r, http.StatusBadRequest, "Title and content are required")
            return
        }

        validIDs, err := app.categoryIDSet()
        if err != nil {
            app.serverError(w, r, "Failed to load categories")
            return
        }

        categoryIDs := make([]int, 0, len(categoryValues))
        seen := make(map[int]struct{})
        for _, value := range categoryValues {
            id, err := strconv.Atoi(value)
            if err != nil || id < 1 {
                continue
            }
            if _, ok := validIDs[id]; !ok {
                continue
            }
            if _, ok := seen[id]; ok {
                continue
            }
            seen[id] = struct{}{}
            categoryIDs = append(categoryIDs, id)
        }

        if len(categoryIDs) == 0 {
            app.clientError(w, r, http.StatusBadRequest, "Select at least one valid category")
            return
        }

        tx, err := app.db.BeginTx(context.Background(), nil)
        if err != nil {
            app.serverError(w, r, "Failed to create post")
            return
        }
        defer tx.Rollback()

        res, err := tx.Exec(
            "INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)",
            user.ID,
            title,
            content,
            time.Now().Unix(),
        )
        if err != nil {
            app.serverError(w, r, "Failed to create post")
            return
        }
        postID64, err := res.LastInsertId()
        if err != nil {
            app.serverError(w, r, "Failed to create post")
            return
        }

        for _, id := range categoryIDs {
            if _, err := tx.Exec(
                "INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
                postID64,
                id,
            ); err != nil {
                app.serverError(w, r, "Failed to assign categories")
                return
            }
        }

        if err := tx.Commit(); err != nil {
            app.serverError(w, r, "Failed to save post")
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
    default:
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
    }
}

func (app *App) postHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/post/")
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) == 0 || parts[0] == "" {
        app.notFound(w, r)
        return
    }

    postID, err := strconv.Atoi(parts[0])
    if err != nil || postID < 1 {
        app.notFound(w, r)
        return
    }

    action := ""
    if len(parts) > 1 {
        action = parts[1]
    }

    switch action {
    case "":
        if r.Method != http.MethodGet {
            app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
            return
        }
        app.viewPost(w, r, postID)
    case "comment":
        if r.Method != http.MethodPost {
            app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
            return
        }
        app.addComment(w, r, postID)
    case "react":
        if r.Method != http.MethodPost {
            app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
            return
        }
        app.reactToPost(w, r, postID)
    default:
        app.notFound(w, r)
    }
}

func (app *App) commentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }
    path := strings.TrimPrefix(r.URL.Path, "/comment/")
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) < 2 {
        app.notFound(w, r)
        return
    }
    commentID, err := strconv.Atoi(parts[0])
    if err != nil || commentID < 1 {
        app.notFound(w, r)
        return
    }
    if parts[1] != "react" {
        app.notFound(w, r)
        return
    }
    app.reactToComment(w, r, commentID)
}

func (app *App) viewPost(w http.ResponseWriter, r *http.Request, postID int) {
    user, err := app.currentUser(w, r)
    if err != nil {
        app.serverError(w, r, "Failed to load session")
        return
    }

    post, err := app.fetchPostByID(postID)
    if errors.Is(err, sql.ErrNoRows) {
        app.notFound(w, r)
        return
    }
    if err != nil {
        app.serverError(w, r, "Failed to load post")
        return
    }

    comments, err := app.fetchComments(postID)
    if err != nil {
        app.serverError(w, r, "Failed to load comments")
        return
    }

    categories, err := app.fetchCategories()
    if err != nil {
        app.serverError(w, r, "Failed to load categories")
        return
    }

    data := TemplateData{
        User:       user,
        Categories: categories,
        Post:       post,
        Comments:   comments,
    }
    app.render(w, "post_view", data, 0)
}

func (app *App) addComment(w http.ResponseWriter, r *http.Request, postID int) {
    user, ok := app.requireAuth(w, r)
    if !ok {
        return
    }
    if err := r.ParseForm(); err != nil {
        app.clientError(w, r, http.StatusBadRequest, "Invalid form")
        return
    }
    content := strings.TrimSpace(r.FormValue("content"))
    if content == "" {
        app.clientError(w, r, http.StatusBadRequest, "Comment cannot be empty")
        return
    }
    if _, err := app.db.Exec(
        "INSERT INTO comments (post_id, user_id, content, created_at) VALUES (?, ?, ?, ?)",
        postID,
        user.ID,
        content,
        time.Now().Unix(),
    ); err != nil {
        app.serverError(w, r, "Failed to add comment")
        return
    }
    http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func (app *App) reactToPost(w http.ResponseWriter, r *http.Request, postID int) {
    user, ok := app.requireAuth(w, r)
    if !ok {
        return
    }
    if err := r.ParseForm(); err != nil {
        app.clientError(w, r, http.StatusBadRequest, "Invalid form")
        return
    }
    value, err := strconv.Atoi(r.FormValue("value"))
    if err != nil || (value != 1 && value != -1) {
        app.clientError(w, r, http.StatusBadRequest, "Invalid reaction")
        return
    }

    if err := app.togglePostReaction(user.ID, postID, value); err != nil {
        app.serverError(w, r, "Failed to react to post")
        return
    }
    http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}

func (app *App) reactToComment(w http.ResponseWriter, r *http.Request, commentID int) {
    user, ok := app.requireAuth(w, r)
    if !ok {
        return
    }
    if err := r.ParseForm(); err != nil {
        app.clientError(w, r, http.StatusBadRequest, "Invalid form")
        return
    }
    value, err := strconv.Atoi(r.FormValue("value"))
    if err != nil || (value != 1 && value != -1) {
        app.clientError(w, r, http.StatusBadRequest, "Invalid reaction")
        return
    }

    if err := app.toggleCommentReaction(user.ID, commentID, value); err != nil {
        app.serverError(w, r, "Failed to react to comment")
        return
    }

    postID, err := app.postIDByComment(commentID)
    if err != nil {
        app.serverError(w, r, "Failed to reload post")
        return
    }
    http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
}
