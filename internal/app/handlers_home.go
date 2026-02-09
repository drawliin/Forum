package app

import (
    "net/http"
    "strconv"
)

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        app.notFound(w, r)
        return
    }
    if r.Method != http.MethodGet {
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    user, err := app.currentUser(w, r)
    if err != nil {
        app.serverError(w, r, "Failed to load session")
        return
    }

    filter := r.URL.Query().Get("filter")
    if filter != "" && filter != "mine" && filter != "liked" {
        app.clientError(w, r, http.StatusBadRequest, "Invalid filter")
        return
    }
    if (filter == "mine" || filter == "liked") && user == nil {
        app.clientError(w, r, http.StatusUnauthorized, "Please log in to use this filter")
        return
    }

    categoryID := 0
    if value := r.URL.Query().Get("category"); value != "" {
        parsed, err := strconv.Atoi(value)
        if err != nil || parsed < 1 {
            app.clientError(w, r, http.StatusBadRequest, "Invalid category")
            return
        }
        categoryID = parsed
    }

    categories, err := app.fetchCategories()
    if err != nil {
        app.serverError(w, r, "Failed to load categories")
        return
    }

    posts, err := app.fetchPosts(user, categoryID, filter)
    if err != nil {
        app.serverError(w, r, "Failed to load posts")
        return
    }

    data := TemplateData{
        User:       user,
        Categories: categories,
        Posts:      posts,
        Filter:     filter,
        CategoryID: categoryID,
    }
    app.render(w, "home", data, 0)
}
