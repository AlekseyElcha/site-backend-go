package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"site-backend-go/internal/middleware"
	"uuid"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func (s *Server) GetAllTicketsInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Context().Value("userID").(uuid.UUID)
	fmt.Println(userID)

	tickets, err := s.TicketService.GetAllTickets(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if tickets == nil || len(tickets) == 0 {
		WriteJSONMessage(w, http.StatusNotFound, "Обращения не найдены.")
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

	WriteJSON(w, http.StatusOK, response)
}

func (s *Server) GetTicketInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	ticketID, err := uuid.Parse(idStr)
	fmt.Println("ticketID:", ticketID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "id is not valid")
		return
	}

	t, err := s.TicketService.GetTicket(ctx, ticketID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			WriteJSONMessage(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var extraMessages []dtos.TicketExtraMessageInfo
	var answerInfo *dtos.TicketAnswerInfoResponse

	answer, err := s.TicketService.GetAnswers(ctx, t.ID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			answer = nil
		} else {
			WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	extraMsgs, err := s.TicketService.GetExtraMessages(ctx, ticketID)
	if err != nil {
		if !errors.Is(err, exceptions.ErrNotFound) {
			WriteError(w, http.StatusInternalServerError, "internal server error")
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

	WriteJSON(w, http.StatusOK, response)
}

func (s *Server) GetTicketsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.PathValue("user_id")
	if userIDStr == "" {
		WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "user_id is not valid")
		return
	}

	tickets, err := s.TicketService.GetAllTicketsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			WriteJSONMessage(w, http.StatusNotFound, "Обращения не найдены.")
			return
		}
		fmt.Println("Реальная ошибка SQL:", err)
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tickets == nil || len(tickets) == 0 {
		WriteError(w, http.StatusNotFound, "Обращения не найдены.")
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

	WriteJSON(w, http.StatusOK, ticketsInfo)
}

func (s *Server) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var t dtos.TicketCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&t)
	if err := validate.Struct(t); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed %s", err.Error()))
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID) // TODO: посмотреть вот это

	err = s.TicketService.Create(ctx, t, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
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

	WriteJSONMessage(w, http.StatusCreated, "Вопрос успешно создан!")
	return
}

func (s *Server) UpdateTicketStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var ts dtos.TicketStatusUpdateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&ts)
	if err := validate.Struct(ts); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err = s.TicketService.UpdateStatus(ctx, ts, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	_, err = s.EmailSenderService.Send(r.Context(),
		[]string{"test@test.com"},
		"Создание обращения",
		"Вы успешно создали обращение на платформе.",
	)
	if err != nil {
		s.Logger.Error(err.Error())
	}

	WriteJSONMessage(w, http.StatusCreated, "Статус вопроса успешно обновлён!")
	return
}

func (s *Server) AnswerTicketByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	senderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var ta dtos.TicketAnswerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&ta)
	if err := validate.Struct(ta); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err = s.TicketService.Answer(ctx, ta, senderID)
	if err != nil {
		fmt.Println("Service error:", err)
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSONMessage(w, http.StatusOK, "Ответ на вопрос успешно создан!")
}

func (s *Server) CreateExtraMessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	senderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var em dtos.ExtraMessageCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&em); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	if err := validate.Struct(em); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err := s.TicketService.CreateExtraMessage(ctx, em, senderID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSONMessage(w, http.StatusCreated, "Дополнительное сообщение успешно создано!")
}

func (s *Server) GetExtraMessagesForTicketByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dtos.ExtraMessagesGetRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	if err := validate.Struct(req); err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	extraMessages, err := s.TicketService.GetExtraMessages(ctx, req.TicketID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = json.NewEncoder(w).Encode(extraMessages)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	WriteJSON(w, http.StatusOK, extraMessages)
}
