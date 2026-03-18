package util

import (
	"forum/internal/models"
	"forum/internal/templates"
	"net/http"
)

func ServerError(w http.ResponseWriter, r *http.Request, message string) {
	redirectError(w, r, http.StatusInternalServerError, message)
}

func ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	redirectError(w, r, status, message)
}

func redirectError(w http.ResponseWriter, r *http.Request, status int, message string) {
	user, _ := CurrentUser(w, r)

	templates.Render(w, "error", models.TemplateData{
		User:      user,
		FormError: message,
		Status:    status,
	}, status)
}
