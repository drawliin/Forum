package app

import (
    "database/sql"
    "errors"
    "net/http"
    "time"

    "github.com/google/uuid"
)

const sessionDuration = 7 * 24 * time.Hour

func (app *App) currentUser(w http.ResponseWriter, r *http.Request) (*User, error) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        return nil, nil
    }

    var user User
    var expires int64
    err = app.db.QueryRow(
        `SELECT u.id, u.email, u.username, s.expires_at
        FROM sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.id = ?`,
        cookie.Value,
    ).Scan(&user.ID, &user.Email, &user.Username, &expires)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    if time.Now().Unix() > expires {
        _, _ = app.db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
        app.clearSessionCookie(w)
        return nil, nil
    }

    return &user, nil
}

func (app *App) requireAuth(w http.ResponseWriter, r *http.Request) (*User, bool) {
    user, err := app.currentUser(w, r)
    if err != nil {
        app.serverError(w, r, "Failed to load session")
        return nil, false
    }
    if user == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return nil, false
    }
    return user, true
}

func (app *App) createSession(w http.ResponseWriter, r *http.Request, userID int) error {
    if _, err := app.db.Exec("DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
        return err
    }

    sessionID := uuid.New().String()
    expires := time.Now().Add(sessionDuration).Unix()

    if _, err := app.db.Exec(
        "INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
        sessionID,
        userID,
        expires,
    ); err != nil {
        return err
    }

    cookie := &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        Expires:  time.Unix(expires, 0),
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    }
    if app.cookieSecure {
        cookie.Secure = true
    }
    http.SetCookie(w, cookie)
    return nil
}

func (app *App) clearSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
    })
}
