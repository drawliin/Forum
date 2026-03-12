package app

import (
	"net/http"
	"os"
)

func (app *App) staticHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.clientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := r.URL.Path[1:]
	info, err := os.Stat(path)
	if err != nil {
		app.clientError(w, r, http.StatusNotFound, "Not found")
		return
	}

	if info.IsDir() {
		app.clientError(w, r, http.StatusNotFound, "Not found")
		return
	}

	http.ServeFile(w, r, path)
}
