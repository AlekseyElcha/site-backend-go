package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"uuid"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func (s *Server) GetAllTicketsInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tickets, err := s.TicketService.GetAllTickets(ctx)
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
			Email:       t.Email,
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
	ctx := r.Context()

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "id is required"}`))
		return
	}

	ticketID, err := uuid.Parse(idStr)
	fmt.Println("ticketID:", ticketID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "id is not valid"}`))
		return
	}

	t, err := s.TicketService.GetTicket(ctx, ticketID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Обращение не найдено."}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	var extraMessages []dtos.TicketExtraMessageInfo
	var answerInfo *dtos.TicketAnswerInfoResponse

	answer, err := s.TicketService.GetAnswers(ctx, t.ID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			answer = nil
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}
	}

	extraMsgs, err := s.TicketService.GetExtraMessages(ctx, ticketID)
	if err != nil {
		if !errors.Is(err, exceptions.ErrNotFound) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
			return
		}
		extraMsgs = nil
	}

	for _, m := range extraMsgs {
		msg := dtos.TicketExtraMessageInfo{
			ID:          m.ID,
			TicketID:    m.TicketID,
			SenderID:    m.SenderID,
			SenderRole:  m.SenderRole,
			MessageText: m.MessageText,
			FilesID:     m.FilesID,
		}

		extraMessages = append(extraMessages, msg)
	}

	if answer != nil {
		answerInfo = &dtos.TicketAnswerInfoResponse{
			ID:          answer.ID,
			CreatedAt:   answer.CreatedAt,
			MessageText: answer.AnswerText,
		}
	} else {
		answerInfo = nil
	}

	response := dtos.TicketGetResponse{
		ID:            t.ID,
		UserID:        t.UserID,
		Name:          t.Name,
		Email:         t.Email,
		Address:       t.Address,
		PhoneNumber:   t.PhoneNumber,
		CreatedAt:     t.CreatedAt,
		Question:      t.Question,
		Status:        t.Status,
		FilesIDs:      t.Files,
		Answer:        answerInfo,
		ExtraMessages: &extraMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetTicketsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	tickets, err := s.TicketService.GetAllTicketsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			http.Error(w, `{"message": "Обращения не найдены."}`, http.StatusNotFound)
			return
		}
		fmt.Println("Реальная ошибка SQL:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal server error`))
		return
	}

	if tickets == nil || len(tickets) == 0 {
		http.Error(w, `{"message": "Обращения не найдены."}`, http.StatusNotFound)
		return
	}

	var ticketsInfo []dtos.TicketGeneralInfo
	for _, ticket := range tickets {
		ticketsInfo = append(ticketsInfo, dtos.TicketGeneralInfo{
			ID:          ticket.ID,
			UserID:      ticket.UserID,
			CreatedAt:   ticket.CreatedAt,
			Name:        ticket.Name,
			Email:       ticket.Email,
			PhoneNumber: ticket.PhoneNumber,
			Address:     ticket.Address,
			Question:    ticket.Question,
			Files:       ticket.Files,
			Status:      ticket.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticketsInfo)
}

func (s *Server) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
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
	ctx := r.Context()

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

	err = s.TicketService.UpdateStatus(ctx, ts, userID)
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
	ctx := r.Context()

	//if r.Method != "POST" {
	//	w.WriteHeader(http.StatusMethodNotAllowed)
	//	w.Write([]byte("Method not allowed"))
	//	return
	//}

	senderID, err := auth.GetUserIDFromCookies(r)
	// senderID, err := uuid.Parse("d3d9d069-c8f6-4c1d-990b-f2c40d209379")

	var ta dtos.TicketAnswerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&ta)
	if err := validate.Struct(ta); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	err = s.TicketService.Answer(ctx, ta, senderID)
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
	ctx := r.Context()

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

	err = s.TicketService.CreateExtraMessage(ctx, em, senderID)
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
	ctx := r.Context()

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

		extraMessages, err := s.TicketService.GetExtraMessages(ctx, req.TicketID)
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
