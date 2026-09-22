package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"time"
	"uuid"
)

type TicketStorage interface {
	GetTicketInfoByID(ctx context.Context, id uuid.UUID) (*db.TicketInfoModel, error)
	GetAllTicketsInfo(ctx context.Context) ([]db.TicketInfoModel, error)
	GetAllTicketsByUserID(ctx context.Context, userID uuid.UUID) ([]db.TicketGeneralInfoModel, error)
	CreateTicket(ctx context.Context, model db.CreateTicketModel) (uuid.UUID, error)
	UpdateTicketStatus(ctx context.Context, model db.UpdateTicketStatusModel) error
	AnswerTicket(ctx context.Context, model db.AnswerTicketModel, senderID uuid.UUID) error
	CreateExtraMessageForTicket(ctx context.Context, model db.ExtraMessageModel) error
	LogTicketOperation(ctx context.Context, model db.TicketLogModel) error
	GetAnswerForTicketByID(ctx context.Context, id uuid.UUID) (*db.TicketAnswersModel, error)
	GetExtraMessagesForTicketByID(ctx context.Context, id uuid.UUID) ([]db.ExtraMessageInfoModel, error)
}
type TicketService struct {
	storage TicketStorage
	log     *slog.Logger
}

func NewTicketService(storage *db.Storage, log *slog.Logger) *TicketService {
	return &TicketService{
		storage: storage,
		log:     log,
	}
}

func (s *TicketService) GetAllTickets(ctx context.Context) ([]db.TicketInfoModel, error) {
	tickets, err := s.storage.GetAllTicketsInfo(ctx)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s *TicketService) GetAllTicketsByUserID(ctx context.Context, userID uuid.UUID) ([]db.TicketGeneralInfoModel, error) {
	tickets, err := s.storage.GetAllTicketsByUserID(ctx, userID)
	fmt.Println(tickets)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s *TicketService) GetTicket(ctx context.Context, id uuid.UUID) (*db.TicketInfoModel, error) {
	t, err := s.storage.GetTicketInfoByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) GetAnswers(ctx context.Context, id uuid.UUID) (*db.TicketAnswersModel, error) {
	t, err := s.storage.GetAnswerForTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		return nil, err
	}

	return t, nil
}

func (s *TicketService) GetExtraMessages(ctx context.Context, id uuid.UUID) ([]db.ExtraMessageInfoModel, error) {
	em, err := s.storage.GetExtraMessagesForTicketByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	return em, nil
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
	}

	if err := s.storage.AnswerTicket(dbCtx, dbModel, senderID); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  ta.TicketID,
		ChangedBy: senderID,
		Action:    "Обновлён статус обращения",
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
