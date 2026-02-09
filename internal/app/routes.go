package app

import (
    "log"
    "net/http"
)

func (app *App) routes() http.Handler {
    mux := http.NewServeMux()
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    mux.HandleFunc("/", app.homeHandler)
    mux.HandleFunc("/register", app.registerHandler)
    mux.HandleFunc("/login", app.loginHandler)
    mux.HandleFunc("/logout", app.logoutHandler)
    mux.HandleFunc("/post/new", app.postNewHandler)
    mux.HandleFunc("/post/", app.postHandler)
    mux.HandleFunc("/comment/", app.commentHandler)
    return mux
}

func (app *App) recoverPanic(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("panic: %v", rec)
                app.serverError(w, r, "Unexpected server error")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
