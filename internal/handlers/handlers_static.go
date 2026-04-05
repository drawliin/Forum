package handlers

import (
	"forum/internal/config"
	"forum/internal/util"
	"net/http"
	"os"
)

func staticHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ClientError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := config.ResolvePath(r.URL.Path[1:])
	info, err := os.Stat(path)
	if err != nil {
		util.ClientError(w, r, http.StatusNotFound, "Not found")
		return
	}

	if info.IsDir() {
		util.ClientError(w, r, http.StatusNotFound, "Not found")
		return
	}

	http.ServeFile(w, r, path)
}
