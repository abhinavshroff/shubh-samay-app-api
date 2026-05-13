package handlers

import (
	"encoding/json"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)
type JSONMap map[string]any

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, JSONMap{"error": message})
}
