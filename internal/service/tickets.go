package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"time"
	"uuid"
)

type TicketStorage interface {
	GetTicketInfoByID(ctx context.Context, id uuid.UUID) (*db.TicketInfoModel, error)
	GetAllTicketsInfo(ctx context.Context) ([]db.TicketGeneralInfoModel, error)
	GetAllTicketsByUserID(ctx context.Context, userID uuid.UUID) ([]db.TicketGeneralInfoModel, error)
	CreateTicket(ctx context.Context, model db.CreateTicketModel) (uuid.UUID, error)
	UpdateTicketStatus(ctx context.Context, model db.UpdateTicketStatusModel) error
	AnswerTicket(ctx context.Context, model db.AnswerTicketModel, senderID uuid.UUID) error
	CreateExtraMessageForTicket(ctx context.Context, model db.ExtraMessageModel) error
	LogTicketOperation(ctx context.Context, model db.TicketLogModel) error
	GetAnswerForTicketByID(ctx context.Context, id uuid.UUID) (*db.TicketAnswersModel, error)
	GetExtraMessagesForTicketByID(ctx context.Context, id uuid.UUID) ([]db.ExtraMessageInfoModel, error)
	SearchForTickets(ctx context.Context, userID string, userRole string, searchingFor string) ([]db.TicketGeneralInfoModel, error)
}
type TicketService struct {
	storage TicketStorage
	log     *slog.Logger
}

func NewTicketService(storage TicketStorage, log *slog.Logger) *TicketService {
	return &TicketService{
		storage: storage,
		log:     log,
	}
}

func (s *TicketService) GetAllTickets(ctx context.Context) ([]dtos.TicketGeneralInfo, error) {
	tickets, err := s.storage.GetAllTicketsInfo(ctx)
	if err != nil {
		return nil, err
	}

	dtosResult := make([]dtos.TicketGeneralInfo, 0, len(tickets))

	for _, ticket := range tickets {
		dtosResult = append(dtosResult, dtos.TicketGeneralInfo{
			ID:        ticket.ID,
			UserID:    ticket.UserID,
			CreatedAt: ticket.CreatedAt,
			Name:      ticket.Name,
			Email:     ticket.Email,
			Address:   ticket.Address,
			Status:    ticket.Status,
		})
	}

	return dtosResult, nil
}

func (s *TicketService) GetAllTicketsByUserID(ctx context.Context, userID uuid.UUID) ([]dtos.TicketGeneralInfo, error) {
	tickets, err := s.storage.GetAllTicketsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var ticketsInfo []dtos.TicketGeneralInfo
	for _, ticket := range tickets {
		ticketsInfo = append(ticketsInfo, dtos.TicketGeneralInfo{
			ID:        ticket.ID,
			UserID:    ticket.UserID,
			CreatedAt: ticket.CreatedAt,
			Name:      ticket.Name,
			Email:     ticket.Email,
			Address:   ticket.Address,
			Status:    ticket.Status,
		})
	}

	return ticketsInfo, nil
}

func (s *TicketService) GetTicket(ctx context.Context, id uuid.UUID) (*dtos.TicketDetailedInfoResponse, error) {
	t, err := s.storage.GetTicketInfoByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dtosResponse := dtos.TicketDetailedInfoResponse{
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

	return &dtosResponse, nil
}

func (s *TicketService) GetAnswer(ctx context.Context, id uuid.UUID) (*dtos.TicketAnswerInfoResponse, error) {
	answer, err := s.storage.GetAnswerForTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		return nil, err
	}

	answerInfo := dtos.TicketAnswerInfoResponse{
		ID:          answer.ID,
		CreatedAt:   answer.CreatedAt,
		MessageText: answer.AnswerText,
		FilesIDs:    answer.Files,
	}
	return &answerInfo, nil
}

func (s *TicketService) GetExtraMessages(ctx context.Context, id uuid.UUID) ([]dtos.TicketExtraMessageInfo, error) {
	em, err := s.storage.GetExtraMessagesForTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	dtosResponse := make([]dtos.TicketExtraMessageInfo, 0, len(em))

	for _, m := range em {
		var attachedFiles []dtos.TicketAttachedFileInfo
		if m.Files != nil {
			attachedFiles = make([]dtos.TicketAttachedFileInfo, 0, len(m.Files))
			for _, file := range m.Files {
				attachedFiles = append(attachedFiles, dtos.TicketAttachedFileInfo{
					FileID:           file.FileID,
					UploaderID:       file.UploaderID,
					OriginalFilename: file.OriginalFilename,
					UploadedAt:       file.UploadedAt,
				})
			}
		} else {
			attachedFiles = make([]dtos.TicketAttachedFileInfo, 0)
		}

		msg := dtos.TicketExtraMessageInfo{
			ID:          m.ID,
			TicketID:    m.TicketID,
			SenderID:    m.SenderID,
			SenderRole:  m.SenderRole,
			MessageText: m.MessageText,
			Files:       attachedFiles,
		}

		dtosResponse = append(dtosResponse, msg)
	}

	return dtosResponse, nil
}

func (s *TicketService) Create(ctx context.Context, t dtos.TicketCreateRequest, userID uuid.UUID) error {
	detachedCtx := context.WithoutCancel(ctx)
	dbCtx, cancel := context.WithTimeout(detachedCtx, 5*time.Second)
	defer cancel()

	dbModel := db.CreateTicketModel{
		UserID:      userID,
		Name:        t.Name,
		Email:       t.Email,
		PhoneNumber: t.PhoneNumber,
		Address:     t.Address,
		Question:    t.Question,
		FilesIDs:    t.FilesIDs,
	}

	ticketID, err := s.storage.CreateTicket(dbCtx, dbModel)
	if err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  ticketID,
		ChangedBy: userID,
		Action:    "Обращение создано",
	}

	err = s.storage.LogTicketOperation(dbCtx, dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) UpdateStatus(ctx context.Context, t dtos.TicketStatusUpdateRequest, userID uuid.UUID) error {
	detachedCtx := context.WithoutCancel(ctx)
	dbCtx, cancel := context.WithTimeout(detachedCtx, 15*time.Second)
	defer cancel()

	dbModel := db.UpdateTicketStatusModel{
		ID:     t.ID,
		Status: t.Status,
	}

	if err := s.storage.UpdateTicketStatus(dbCtx, dbModel); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  t.ID,
		ChangedBy: userID,
		Action:    "Обновлён статус обращения",
	}

	err := s.storage.LogTicketOperation(dbCtx, dbLogModel)
	if err != nil {
		slog.ErrorContext(ctx, "ticket status updated, but failed to log operation",
			"ticket_id", t.ID,
			"status", t.Status,
			"error", err,
		)
		return err
	}

	return nil
}

func (s *TicketService) Answer(ctx context.Context, ta dtos.TicketAnswerRequest, senderID uuid.UUID) error {
	detachedCtx := context.WithoutCancel(ctx)
	dbCtx, cancel := context.WithTimeout(detachedCtx, 15*time.Second)
	defer cancel()

	dbModel := db.AnswerTicketModel{
		ID:         ta.TicketID,
		AnswerText: ta.AnswerText,
		FilesIDs:   &ta.FilesIDs,
	}

	if err := s.storage.AnswerTicket(dbCtx, dbModel, senderID); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  ta.TicketID,
		ChangedBy: senderID,
		Action:    "Отправлен ответ на обращение",
	}

	err := s.storage.LogTicketOperation(dbCtx, dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) CreateExtraMessage(ctx context.Context, em dtos.ExtraMessageCreateRequest, senderID uuid.UUID) error {
	detachedCtx := context.WithoutCancel(ctx)
	dbCtx, cancel := context.WithTimeout(detachedCtx, 15*time.Second)
	defer cancel()

	dbModel := db.ExtraMessageModel{
		TicketID:    em.TicketID,
		SenderID:    senderID,
		MessageText: em.MessageText,
		FilesID:     em.FilesIDs,
	}

	if err := s.storage.CreateExtraMessageForTicket(dbCtx, dbModel); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  em.TicketID,
		ChangedBy: senderID,
		Action:    "Создано дополнительное сообщение",
	}

	err := s.storage.LogTicketOperation(dbCtx, dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) FindSuitableTicketsFilter(ctx context.Context, userID string, userRole string, searchFor string) ([]dtos.TicketGeneralInfo, error) {
	detachedCtx := context.WithoutCancel(ctx)
	dbCtx, cancel := context.WithTimeout(detachedCtx, 15*time.Second)
	defer cancel()

	ticketsFound, err := s.storage.SearchForTickets(dbCtx, userID, userRole, searchFor)
	if err != nil {
		return nil, err
	}

	dtosResult := make([]dtos.TicketGeneralInfo, 0, len(ticketsFound))

	for _, ticket := range ticketsFound {
		dtosResult = append(dtosResult, dtos.TicketGeneralInfo{
			ID:        ticket.ID,
			UserID:    ticket.UserID,
			CreatedAt: ticket.CreatedAt,
			Name:      ticket.Name,
			Email:     ticket.Email,
			Address:   ticket.Address,
			Status:    ticket.Status,
		})
	}

	return dtosResult, nil
}
