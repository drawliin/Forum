package util

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"forum/internal/config"
	"forum/internal/db"
	"forum/internal/models"

	"github.com/google/uuid"
)

const sessionDuration = 7 * 24 * time.Hour

func CurrentUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, nil
	}

	var user models.User
	var expires int64
	err = db.Database.QueryRow(
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
		_, err = db.Database.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		if err != nil {
			return nil, err
		}

		ClearSessionCookie(w)
		return nil, nil
	}

	return &user, nil
}

func RequireAuth(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	user, err := CurrentUser(w, r)
	if err != nil {
		ServerError(w, r, "Failed to load session")
		return nil, false
	}
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}
	return user, true
}

func CreateSession(w http.ResponseWriter, r *http.Request, userID int) error {
	if _, err := db.Database.Exec("DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
		return err
	}

	sessionID := uuid.New().String()
	expires := time.Now().Add(sessionDuration).Unix()

	if _, err := db.Database.Exec(
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
	if config.GetConfig().CookieSecure {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
	return nil
}

func ClearSessionCookie(w http.ResponseWriter) {
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
