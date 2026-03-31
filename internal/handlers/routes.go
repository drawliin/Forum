package handlers

import (
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
