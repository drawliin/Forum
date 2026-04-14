package util

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, likes, dislikes int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}
