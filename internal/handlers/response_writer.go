package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.Logger.Error("WriteJSON", "error", err)
	}
}

func (s *Server) WriteJSONMessage(w http.ResponseWriter, statusCode int, message string) {
	s.WriteJSON(w, statusCode, map[string]string{
		"message": message},
	)
}

func (s *Server) WriteError(w http.ResponseWriter, statusCode int, message string) {
	s.WriteJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
