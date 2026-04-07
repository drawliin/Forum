package util

import (
	"forum/internal/models"
	"forum/internal/templates"
	"net/http"
)

// ServerError renders the shared error page with a 500 status.
func ServerError(w http.ResponseWriter, r *http.Request, message string) {
	redirectError(w, r, http.StatusInternalServerError, message)
}

// ClientError renders the shared error page for bad requests and missing pages.
func ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	redirectError(w, r, status, message)
}

// redirectError keeps error responses in one place so handlers stay small.
func redirectError(w http.ResponseWriter, r *http.Request, status int, message string) {
	user, _ := CurrentUser(w, r)

	templates.Render(w, "error", models.TemplateData{
		User:      user,
		FormError: message,
		Status:    status,
	}, status)
}
