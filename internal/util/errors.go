package util

import (
	"forum/internal/models"
	"forum/internal/templates"
	"net/http"
)

func ServerError(w http.ResponseWriter, r *http.Request, message string) {
	user, _ := CurrentUser(w, r)
	templates.Render(w, "error", models.TemplateData{
		User:      user,
		FormError: message,
		Status:    http.StatusInternalServerError,
	}, http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	user, _ := CurrentUser(w, r)
	templates.Render(w, "error", models.TemplateData{
		User:      user,
		FormError: message,
		Status:    status,
	}, status)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	ClientError(w, r, http.StatusNotFound, "Page not found")
}
