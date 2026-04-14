package handlers

import (
	"forum/internal/util"
	"net/http"
	"sync"
	"time"
)

const maxTokens = 1000
const refillTime = time.Second

var tokens = maxTokens
var lastRefill = time.Now()
var mutex sync.Mutex

// SetupRoutes connects URL paths to their handlers.
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
	return rateLimiterMiddleware(mux)
}

// rateLimiterMiddleware slows down bursts of requests with a simple token bucket.
func rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		elapsed := time.Since(lastRefill)
		if elapsed > refillTime {
			// refill tokens
			tokens += min(int(elapsed/refillTime), maxTokens)
			lastRefill = time.Now()
		}

		if tokens == 0 {
			util.ClientError(w, r, http.StatusTooManyRequests, "Too many requests")
		} else {
			tokens -= 1
			next.ServeHTTP(w, r)
		}
	})
}
