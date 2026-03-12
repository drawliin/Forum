package app

import "net/http"

func (app *App) serverError(w http.ResponseWriter, r *http.Request, message string) {
	user, _ := app.currentUser(w, r)
	app.render(w, "error", TemplateData{
		User:      user,
		FormError: message,
		Status:    http.StatusInternalServerError,
	}, http.StatusInternalServerError)
}

func (app *App) clientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	user, _ := app.currentUser(w, r)
	app.render(w, "error", TemplateData{
		User:      user,
		FormError: message,
		Status:    status,
	}, status)
}

func (app *App) notFound(w http.ResponseWriter, r *http.Request) {
	app.clientError(w, r, http.StatusNotFound, "Page not found")
}
