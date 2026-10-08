package handlers

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// log
	}
}

func WriteJSONMessage(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]string{
		"message": message},
	)
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
