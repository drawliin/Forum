package app

import (
    "database/sql"
    "errors"
    "net/http"
    "strings"
    "time"

    "golang.org/x/crypto/bcrypt"
)

func (app *App) registerHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        user, _ := app.currentUser(w, r)
        app.render(w, "register", TemplateData{User: user}, 0)
        return
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            app.clientError(w, r, http.StatusBadRequest, "Invalid form")
            return
        }
        email := strings.TrimSpace(r.FormValue("email"))
        username := strings.TrimSpace(r.FormValue("username"))
        password := r.FormValue("password")

        if email == "" || username == "" || password == "" {
            app.clientError(w, r, http.StatusBadRequest, "All fields are required")
            return
        }

        var existing int
        err := app.db.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", email, username).Scan(&existing)
        if err == nil {
            app.clientError(w, r, http.StatusBadRequest, "Email or username already taken")
            return
        }
        if !errors.Is(err, sql.ErrNoRows) {
            app.serverError(w, r, "Failed to validate user")
            return
        }

        hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            app.serverError(w, r, "Failed to secure password")
            return
        }

        _, err = app.db.Exec(
            "INSERT INTO users (email, username, password_hash, created_at) VALUES (?, ?, ?, ?)",
            email,
            username,
            string(hash),
            time.Now().Unix(),
        )
        if err != nil {
            app.serverError(w, r, "Failed to create user")
            return
        }

        http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
    default:
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
    }
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        user, _ := app.currentUser(w, r)
        info := ""
        if r.URL.Query().Get("registered") == "1" {
            info = "Account created. You can log in now."
        }
        app.render(w, "login", TemplateData{User: user, Info: info}, 0)
        return
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            app.clientError(w, r, http.StatusBadRequest, "Invalid form")
            return
        }
        email := strings.TrimSpace(r.FormValue("email"))
        password := r.FormValue("password")

        if email == "" || password == "" {
            app.clientError(w, r, http.StatusBadRequest, "Email and password are required")
            return
        }

        var user User
        var hash string
        err := app.db.QueryRow(
            "SELECT id, username, email, password_hash FROM users WHERE email = ?",
            email,
        ).Scan(&user.ID, &user.Username, &user.Email, &hash)
        if errors.Is(err, sql.ErrNoRows) {
            app.clientError(w, r, http.StatusUnauthorized, "Invalid credentials")
            return
        }
        if err != nil {
            app.serverError(w, r, "Failed to authenticate")
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
            app.clientError(w, r, http.StatusUnauthorized, "Invalid credentials")
            return
        }

        if err := app.createSession(w, r, user.ID); err != nil {
            app.serverError(w, r, "Failed to create session")
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
    default:
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
    }
}

func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }
    cookie, err := r.Cookie("session_id")
    if err == nil {
        _, _ = app.db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
    }
    app.clearSessionCookie(w)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
