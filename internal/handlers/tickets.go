package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"site-backend-go/internal/config"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"site-backend-go/internal/middleware"
	"uuid"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func (s *Server) GetAllTicketsInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tickets, err := s.TicketService.GetAllTickets(ctx)
	if err != nil {
		if errors.Is(err, exceptions.ErrUnauthorized) {
			s.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if errors.Is(err, exceptions.ErrForbidden) {
			s.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if tickets == nil || len(tickets) == 0 {
		s.WriteJSONMessage(w, http.StatusNotFound, "Обращения не найдены.")
		return
	}

	response := dtos.ItemResponse[dtos.TicketGeneralInfo]{
		Items: tickets,
	}
	s.WriteJSON(w, http.StatusOK, response)
}

func (s *Server) GetTicketInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		s.WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	ticketID, err := uuid.Parse(idStr)
	fmt.Println("ticketID:", ticketID)
	if err != nil {
		s.WriteError(w, http.StatusBadRequest, "id is not valid")
		return
	}

	ticket, err := s.TicketService.GetTicket(ctx, ticketID)
	if err != nil {
		if errors.Is(err, exceptions.ErrForbidden) {
			s.WriteJSON(w, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, exceptions.ErrNotFound) {
			s.WriteJSONMessage(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	filesIDs := ticket.FilesIDs
	ticket.FilesIDs = nil

	answer, err := s.TicketService.GetAnswer(ctx, ticket.ID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			answer = nil
		}
	}

	extraMsgs, err := s.TicketService.GetExtraMessages(ctx, ticketID)
	if err != nil {
		if !errors.Is(err, exceptions.ErrNotFound) {
			s.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		extraMsgs = nil
	}

	files, err := s.FileService.GetMultipleFilesInfo(ctx, filesIDs)
	if err != nil {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		fmt.Printf(err.Error())
		return
	}

	filesResponse := dtos.ItemResponse[dtos.FileDetailsInfo]{
		Items: files,
	}

	extraMsgsResponse := dtos.ItemResponse[dtos.TicketExtraMessageInfo]{
		Items: extraMsgs,
	}

	response := dtos.TicketGetResponse{
		Info:          ticket,
		Files:         &filesResponse,
		ExtraMessages: &extraMsgsResponse,
		Answers:       answer,
	}

	s.WriteJSON(w, http.StatusOK, response)
}

func (s *Server) GetTicketsByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requesterID, okID := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	requesterRole, okRole := ctx.Value(middleware.UserRoleKey).(string)
	if !okID || !okRole {
		s.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDStr := r.PathValue("user_id")
	if userIDStr == "" {
		s.WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	requesterIDStr := requesterID.String()

	if requesterIDStr != userIDStr && requesterRole != "admin" {
		s.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.WriteError(w, http.StatusBadRequest, "user_id is not valid")
		return
	}

	tickets, err := s.TicketService.GetAllTicketsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			s.WriteJSONMessage(w, http.StatusNotFound, "Обращения не найдены.")
			return
		}
		fmt.Println("Реальная ошибка SQL:", err)
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tickets == nil || len(tickets) == 0 {
		s.WriteError(w, http.StatusNotFound, "Обращения не найдены.")
		return
	}

	response := dtos.ItemResponse[dtos.TicketGeneralInfo]{
		Items: tickets,
	}

	s.WriteJSON(w, http.StatusOK, response)
}

func (s *Server) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var t dtos.TicketCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&t)
	if err := validate.Struct(t); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed %s", err.Error()))
		return
	}

	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	err = s.TicketService.Create(ctx, t, userID)
	if err != nil {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	userEmail := t.Email
	formattedUserMessage := fmt.Sprintf(config.TicketCreatedMessageForUser, t.Name)
	formattedAdminMessage := fmt.Sprintf(config.TicketCreatedMessageForAdmins, t.Email, t.Name, t.Address)

	go func() {
		emailIDUser, errUser := s.EmailSenderService.Send(
			ctx,
			[]string{userEmail},
			config.TicketCreatedSubjectForUser,
			formattedUserMessage,
		)
		if errUser != nil {
			s.Logger.Error("error sending user email for ticket creation", "error", errUser.Error())
		}

		s.Logger.Info("email sent to user for ticket creation", "email", emailIDUser)
	}()

	go func() {
		emailIDAdmin, errAdmin := s.EmailSenderService.Send(
			ctx,
			config.AdminEmails,
			config.TicketCreatedSubjectForAdmins,
			formattedAdminMessage,
		)
		if errAdmin != nil {
			s.Logger.Error("error sending admin email for ticket creation", "error", errAdmin.Error())
		}

		s.Logger.Info("email sent to admins for ticket creation", "email", emailIDAdmin)
	}()

	s.WriteJSONMessage(w, http.StatusCreated, "Вопрос успешно создан!")
	return
}

func (s *Server) UpdateTicketStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		s.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var ts dtos.TicketStatusUpdateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&ts)
	if err := validate.Struct(ts); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err = s.TicketService.UpdateStatus(ctx, ts, userID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			s.WriteError(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	//if err != nil {
	//	s.Logger.Error(err.Error())
	//}

	s.WriteJSONMessage(w, http.StatusCreated, "Статус вопроса успешно обновлён!")
	return
}

func (s *Server) AnswerTicketByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	senderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var ta dtos.TicketAnswerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&ta)
	if err := validate.Struct(ta); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err = s.TicketService.Answer(ctx, ta, senderID)
	if err != nil {
		fmt.Println("Service error:", err)
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	ticketInfo, err := s.TicketService.GetTicket(ctx, ta.TicketID)
	if err != nil {
		s.Logger.Error("error getting ticket info", "error", err)
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	userEmail := ticketInfo.Email

	formattedUserMessage := fmt.Sprintf(config.TicketAnsweredMessageForUser, ticketInfo.Name)
	emailID, err := s.EmailSenderService.Send(
		ctx,
		[]string{userEmail},
		config.TicketAnsweredSubjectForUser,
		formattedUserMessage,
	)

	s.Logger.Info("email sent to user for ticket update", "email", emailID, "ticketID", ticketInfo.ID)

	s.WriteJSONMessage(w, http.StatusOK, "Ответ на вопрос успешно создан!")
}

func (s *Server) CreateExtraMessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	senderID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var em dtos.ExtraMessageCreateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&em); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	if err := validate.Struct(em); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	err := s.TicketService.CreateExtraMessage(ctx, em, senderID)
	if err != nil {
		if errors.Is(err, exceptions.ErrNotFound) {
			s.WriteError(w, http.StatusNotFound, "Обращение не найдено.")
			return
		}
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	s.WriteJSONMessage(w, http.StatusCreated, "Дополнительное сообщение успешно создано!")
}

func (s *Server) GetExtraMessagesForTicketByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dtos.ExtraMessagesGetRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	if err := validate.Struct(req); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	extraMessages, err := s.TicketService.GetExtraMessages(ctx, req.TicketID)

	if err != nil {
		if errors.Is(err, exceptions.ErrUnauthorized) {
			s.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if errors.Is(err, exceptions.ErrForbidden) {
			s.WriteError(w, http.StatusForbidden, "forbidden")
			return
		}

		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = json.NewEncoder(w).Encode(extraMessages)
	if err != nil {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	s.WriteJSON(w, http.StatusOK, extraMessages)
}

func (s *Server) SearchTicketsAlikeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		s.WriteError(w, http.StatusInternalServerError, "authentication issue")
		return
	}

	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	ctx := r.Context()

	queryParams := r.URL.Query()
	searchFor := queryParams.Get("search")
	searchForInput := dtos.TicketAlikeSearch{
		InputData: searchFor,
	}

	if err := validate.Struct(searchForInput); err != nil {
		s.WriteError(w, http.StatusBadRequest, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	userIDStr := userID.String()
	ticketsFound, err := s.TicketService.FindSuitableTicketsFilter(ctx, userIDStr, userRole, searchForInput.InputData)
	if err != nil {
		if errors.Is(err, exceptions.ErrDBError) {
			s.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
	}

	response := dtos.ItemResponse[dtos.TicketGeneralInfo]{
		Items: ticketsFound,
	}
	s.WriteJSON(w, http.StatusOK, response)
}
