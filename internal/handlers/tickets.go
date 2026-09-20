package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/config"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
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
	Config             config.Config
}

func NewServer(
	ticketService *service.TicketService,
	userService *service.UserService,
	emailSenderService service.EmailSenderService,
	fileService *service.FileService,
	config *config.Config,
) *Server {
	return &Server{
		TicketService:      *ticketService,
		UserService:        *userService,
		EmailSenderService: emailSenderService,
		FileService:        *fileService,
		Config:             *config,
	}
}

func (s *Server) GetAllTicketsInfoHandler(w http.ResponseWriter, r *http.Request) {
	tickets, err := s.TicketService.GetAllTickets()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"`))
		return
	}
	if tickets == nil || len(tickets) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":  "Обращения не найдены."}`))
		return
	}

	response := make([]dtos.TicketGetResponse, len(tickets))
	for i, t := range tickets {
		response[i] = dtos.TicketGetResponse{
			ID:          t.ID,
			UserID:      t.UserID,
			Name:        t.Name,
			Address:     t.Address,
			PhoneNumber: t.PhoneNumber,
			CreatedAt:   t.CreatedAt,
			Question:    t.Question,
			Status:      t.Status,
			FilesIDs:    t.Files,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Обращение не найдено."))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	response := dtos.TicketGetResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Name:        t.Name,
		Address:     t.Address,
		PhoneNumber: t.PhoneNumber,
		CreatedAt:   t.CreatedAt,
		Question:    t.Question,
		Status:      t.Status,
		FilesIDs:    t.Files,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetTicketsByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "user_id is invalid", http.StatusBadRequest)
		return
	}

	tickets, err := s.TicketService.GetAllTicketsByUserID(userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			http.Error(w, `{"message": "Обращения не найдены."}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal server error`))
	}

	if tickets == nil || len(tickets) == 0 {
		http.Error(w, `{"message": "Обращения не найдены."}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
}

func (s *Server) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	ctx := r.Context()

	userID, err := auth.GetUserIDFromCookies(r)
	if err != nil {
		fmt.Println(err)
		fmt.Println(userID)
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

	err = s.TicketService.Create(ctx, t, userID)
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
	w.Write([]byte(`{"message":  "Вопрос успешно создан!"}`))
	return
}

func (s *Server) UpdateTicketStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	userID, err := auth.GetUserIDFromCookies(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}

	var ts dtos.TicketStatusUpdateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&ts)
	if err := validate.Struct(ts); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.UpdateStatus(ts, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Обращение не найдено."))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
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
	w.Write([]byte(`{"message":  "Статус вопроса успешно обновлён!"}`))
	return
}

func (s *Server) AnswerTicketByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	senderID, err := auth.GetUserIDFromCookies(r)
	//senderID, err := uuid.Parse("b446a1e6-898e-4e27-aaba-f55c4869cafc")

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
	w.Write([]byte(`{"message":  "Ответ на вопрос успешно создан!"}`))
}

func (s *Server) CreateExtraMessageHandler(w http.ResponseWriter, r *http.Request) {
	senderID, err := auth.GetUserIDFromCookies(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	var em dtos.ExtraMessageCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&em); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := validate.Struct(em); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.CreateExtraMessage(em, senderID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Обращение не найдено."))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Дополнительное сообщение успешно создано!"}`))
}

func (s *Server) GetExtraMessagesForTicketByID(w http.ResponseWriter, r *http.Request) {
	{
		var req dtos.ExtraMessagesGetRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if err := validate.Struct(req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
			return
		}

		extraMessages, err := s.TicketService.GetExtraMessages(req.TicketID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(extraMessages)
		if err != nil {
			w.Write([]byte(`{"error":"internal server error"}`))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
