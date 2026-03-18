package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"
	"net/http"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := util.CurrentUser(w, r)
	if err != nil {
		util.ServerError(w, r, "Failed to get current user")
		return
	}

	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		templates.Render(w, "register", models.TemplateData{}, 0)
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			util.ClientError(w, r, http.StatusBadRequest, "Invalid form")
			return
		}
		email := strings.TrimSpace(r.FormValue("email"))
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		if err = validateInput(username, email, password); err != nil {
			util.ClientError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		var existing int
		err := db.Database.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", email, username).Scan(&existing)
		if err == nil {
			util.ClientError(w, r, http.StatusBadRequest, "Email or username already taken")
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			util.ServerError(w, r, "Failed to validate user")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			util.ServerError(w, r, "Failed to secure password")
			return
		}

		_, err = db.Database.Exec(
			"INSERT INTO users (email, username, password_hash, created_at) VALUES (?, ?, ?, ?)",
			email,
			username,
			string(hash),
			time.Now().Unix(),
		)
		if err != nil {
			util.ServerError(w, r, "Failed to create user")
			return
		}

		http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
	default:
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := util.CurrentUser(w, r)
	if err != nil {
		util.ServerError(w, r, "Failed to get current user")
		return
	}

	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		info := ""
		if r.URL.Query().Get("registered") == "1" {
			info = "Account created. You can log in now."
		}
		templates.Render(w, "login", models.TemplateData{User: user, Info: info}, 0)
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			util.ClientError(w, r, http.StatusBadRequest, "Invalid form")
			return
		}
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")

		if email == "" || password == "" {
			util.ClientError(w, r, http.StatusBadRequest, "Email and password are required")
			return
		}

		var user models.User
		var hash string
		err := db.Database.QueryRow(
			"SELECT id, username, email, password_hash FROM users WHERE email = ?",
			email,
		).Scan(&user.ID, &user.Username, &user.Email, &hash)
		if errors.Is(err, sql.ErrNoRows) {
			util.ClientError(w, r, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		if err != nil {
			util.ServerError(w, r, "Failed to authenticate")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			util.ClientError(w, r, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		if err := util.CreateSession(w, r, user.ID); err != nil {
			util.ServerError(w, r, "Failed to create session")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: check that user is logged in

	cookie, err := r.Cookie("session_id")
	if err == nil {
		_, _ = db.Database.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	}
	util.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func validateInput(username string, email string, password string) error {
	if len(username) == 0 || len(email) == 0 || len(password) == 0 {
		return fmt.Errorf("All fields are required")
	}

	for _, r := range username + email + password {
		if unicode.IsSpace(r) {
			return fmt.Errorf("Fields cannot contain whitespaces")
		}
	}

	split := strings.Split(email, "@")
	if len(split) != 2 || len(split[0]) == 0 || len(split[1]) == 0 {
		return fmt.Errorf("Invalid email format")
	}

	return nil
}
