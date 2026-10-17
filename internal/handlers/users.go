package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/middleware"
	"uuid"
)

func (s *Server) CreateNewUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var u dtos.UserCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON format", "details":"` + err.Error() + `"}`))
		return
	}

	if err := validate.Struct(u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	_, err := s.UserService.Create(ctx, u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(""))
	return
}

func (s *Server) UpdateUserInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var u dtos.UserUpdateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Ошибка декодирования данных."}`))
		return
	}

	if err := validate.Struct(u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	targetUserID := u.ID
	requesterID, okID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	requesterRole, okRole := ctx.Value(middleware.UserRoleKey).(string)
	if !okID || !okRole {
		s.WriteError(w, http.StatusUnauthorized, "unauthorized")
		fmt.Println(okID)
		return
	}

	if targetUserID != requesterID && requesterRole != "admin" {
		s.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	s.UserService.Update(ctx, u)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Данные аккаунта успешно обновлены!"}`))
}

func (s *Server) GetUserInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requesterID, okID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	requesterRole, okRole := ctx.Value(middleware.UserRoleKey).(string)
	if !okID || !okRole {
		s.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requesterIDStr := requesterID.String()

	idStr := r.PathValue("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "id is required"}`))
		return
	}

	if idStr != requesterIDStr && requesterRole != "admin" {
		s.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	userInfo, err := s.UserService.GetUserByID(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userInfo); err != nil {
		log.Printf("failed to encode user response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}
	return
}
