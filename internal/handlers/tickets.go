package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/service"
	"uuid"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Server struct {
	TicketService      service.TicketService
	UserService        service.UserService
	EmailSenderService service.EmailSenderService
	FileService        service.FileService
}

func NewServer(
	ticketService service.TicketService,
	userService service.UserService,
	emailSenderService service.EmailSenderService,
	FileService service.FileService,
) *Server {
	return &Server{
		TicketService:      ticketService,
		UserService:        userService,
		EmailSenderService: emailSenderService,
		FileService:        FileService,
	}
}

func (s *Server) GetTicketInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	ticketID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "id is invalid", http.StatusBadRequest)
		return
	}

	t, err := s.TicketService.GetTicket(ticketID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	response := dtos.TicketGetResponse{
		ID:        t.ID,
		CreatedAt: t.CreatedAt,
		Question:  t.Question,
		Status:    t.Status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return

	}

	userID, err := auth.GetUserIDFromCookies(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	var t dtos.TicketCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&t)
	if err := validate.Struct(t); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.Create(t, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	userEmail := t.Email

	emailID, err := s.EmailSenderService.Send(
		r.Context(),
		[]string{userEmail},
		"Создание обращения",
		"<h1>Статус изменен</h1><p>Ваше обращение успешно обновлено!</p>",
	)
	if err != nil {
		fmt.Printf("[ОШИБКА EMAIL] Не удалось отправить письмо для %s: %v\n", emailID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Ticket created"))
	return
}

func (s *Server) UpdateTicketStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	var ts dtos.TicketStatusUpdateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&ts)
	if err := validate.Struct(ts); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.UpdateStatus(ts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) // переписать
		return
	}

	_, err = s.EmailSenderService.Send(r.Context(),
		[]string{"test@test.com"},
		"Создание обращения",
		"Вы успешно создали обращение на платформе.",
	)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Ticket updated"))
	return
}

func (s *Server) AnswerTicketByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	// senderID, err := auth.GetUserIDFromCookies(r)
	senderID, err := uuid.Parse("b446a1e6-898e-4e27-aaba-f55c4869cafc")

	var ta dtos.TicketAnswerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&ta)
	if err := validate.Struct(ta); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.Answer(ta, senderID)
	if err != nil {
		fmt.Println("Service error:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("answered"))
}

func (s *Server) CreateNewUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	var u dtos.UserCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&u)
	if err := validate.Struct(u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.UserService.Create(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) // переписать
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created"))
	return
}
