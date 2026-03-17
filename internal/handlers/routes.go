package handlers

import (
	"forum/internal/util"
	"log"
	"net/http"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/post/new", postNewHandler)
	mux.HandleFunc("/post/", postHandler)
	mux.HandleFunc("/comment/", commentHandler)
	return mux
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				util.ServerError(w, r, "Unexpected server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
