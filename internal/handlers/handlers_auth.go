package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"forum/internal/db"
	"forum/internal/models"
	"forum/internal/templates"
	"forum/internal/util"

	"golang.org/x/crypto/bcrypt"
)

// registerHandler shows the register page and creates new accounts.
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
		email = strings.ToLower(email)
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		// Trying to fix
		if err = validateInput(username, email, password); err != nil {
			// Problem with email/username/password
			templates.Render(w, "register", models.TemplateData{
				FormError: err.Error(),
			}, http.StatusBadRequest)

			return
		}

		var existing int
		err := db.Database.QueryRow("SELECT id FROM users WHERE email = ? OR username = ?", email, username).Scan(&existing)
		if err == nil {
			// Email or username already taken
			templates.Render(w, "register", models.TemplateData{
				FormError: "Email or username already taken",
			}, http.StatusBadRequest)
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

// loginHandler checks credentials and creates a session after a successful login.
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
		email = strings.ToLower(email)
		password := r.FormValue("password")

		if email == "" || password == "" {
			// No Email or Password
			templates.Render(w, "login", models.TemplateData{
				FormError: "Email and password are required",
			}, http.StatusBadRequest)
			return
		}

		var user models.User
		var hash string
		err := db.Database.QueryRow(
			"SELECT id, username, email, password_hash FROM users WHERE email = ?",
			email,
		).Scan(&user.ID, &user.Username, &user.Email, &hash)
		if errors.Is(err, sql.ErrNoRows) {

			// Invalid credentials: email doesn't exist
			templates.Render(w, "login", models.TemplateData{
				FormError: "Invalid credentials",
			}, http.StatusBadRequest)
			return
		}
		if err != nil {
			util.ServerError(w, r, "Failed to authenticate")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
			// Invalid credentials: wrong password
			templates.Render(w, "login", models.TemplateData{
				FormError: "Invalid credentials",
			}, http.StatusBadRequest)
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

// logoutHandler removes the current session and sends the user back home.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	_, err := util.CurrentUser(w, r)
	if err != nil {
		// not logged in
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil {
		_, _ = db.Database.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	}
	util.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// validateInput does the basic checks for register form fields.
func validateInput(username string, email string, password string) error {
	if len(username) > 30 || len(email) > 30 || len(password) > 64 {
		return fmt.Errorf("field too long")
	}

	if len(username) == 0 || len(email) == 0 || len(password) == 0 {
		return fmt.Errorf("all fields are required")
	}

	if len(password) < 6 {
		return fmt.Errorf("password too short")
	}

	for _, r := range username + email + password {
		if unicode.IsSpace(r) {
			return fmt.Errorf("fields cannot contain whitespaces")
		}
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
