package service

import (
	"context"
	"database/sql"
	"errors"
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"uuid"
)

type TicketStorage interface {
	GetTicketInfoByID(id uuid.UUID) (*db.TicketInfoModel, error)
	GetAllTicketsInfo() ([]db.TicketInfoModel, error)
	GetAllTicketsByUserID(userID uuid.UUID) ([]db.TicketInfoModel, error)
	CreateTicket(ctx context.Context, model db.CreateTicketModel) (uuid.UUID, error)
	UpdateTicketStatus(model db.UpdateTicketStatusModel) error
	AnswerTicket(model db.AnswerTicketModel, senderID uuid.UUID) error
	CreateExtraMessageForTicket(model db.ExtraMessageModel) error
	LogTicketOperation(model db.TicketLogModel) error
	GetAnswersForTicketByID(id uuid.UUID) ([]db.TicketAnswersModel, error)
	GetExtraMessagesForTicketByID(id uuid.UUID) ([]db.ExtraMessageInfoModel, error)
}
type TicketService struct {
	storage TicketStorage
}

func NewTicketService(storage TicketStorage) *TicketService {
	return &TicketService{
		storage: storage,
	}
}

func (s *TicketService) GetAllTickets() ([]db.TicketInfoModel, error) {
	tickets, err := s.storage.GetAllTicketsInfo()
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s *TicketService) GetAllTicketsByUserID(userID uuid.UUID) ([]db.TicketInfoModel, error) {
	tickets, err := s.storage.GetAllTicketsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s *TicketService) GetTicket(id uuid.UUID) (*db.TicketInfoModel, error) {
	t, err := s.storage.GetTicketInfoByID(id)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) GetAnswers(id uuid.UUID) ([]db.TicketAnswersModel, error) {
	t, err := s.storage.GetAnswersForTicketByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return t, err
}

func (s *TicketService) GetExtraMessages(id uuid.UUID) ([]db.ExtraMessageInfoModel, error) {
	em, err := s.storage.GetExtraMessagesForTicketByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	return em, nil
}

func (s *TicketService) Create(ctx context.Context, t dtos.TicketCreateRequest, userID uuid.UUID) error {
	dbModel := db.CreateTicketModel{
		UserID:      userID,
		Name:        t.Name,
		Email:       t.Email,
		PhoneNumber: t.PhoneNumber,
		Address:     t.Address,
		Question:    t.Question,
	}

	type TicketLogModel struct {
		TicketID  uuid.UUID
		ChangedBy uuid.UUID
		Action    string
	}

	ticketID, err := s.storage.CreateTicket(ctx, dbModel)
	if err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  ticketID,
		ChangedBy: userID,
		Action:    "Обращение создано",
	}

	err = s.storage.LogTicketOperation(dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) UpdateStatus(t dtos.TicketStatusUpdateRequest, userID uuid.UUID) error {
	dbModel := db.UpdateTicketStatusModel{
		ID:     t.ID,
		Status: t.Status,
	}

	if err := s.storage.UpdateTicketStatus(dbModel); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  t.ID,
		ChangedBy: userID,
		Action:    "Обновлён статус обращения",
	}

	err := s.storage.LogTicketOperation(dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) Answer(ta dtos.TicketAnswerRequest, senderID uuid.UUID) error {
	dbModel := db.AnswerTicketModel{
		ID:         ta.TicketID,
		AnswerText: ta.AnswerText,
	}

	if err := s.storage.AnswerTicket(dbModel, senderID); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  ta.TicketID,
		ChangedBy: senderID,
		Action:    "Обновлён статус обращения",
	}

	err := s.storage.LogTicketOperation(dbLogModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) CreateExtraMessage(em dtos.ExtraMessageCreateRequest, senderID uuid.UUID) error {
	dbModel := db.ExtraMessageModel{
		TicketID:    em.TicketID,
		SenderID:    senderID,
		MessageText: em.MessageText,
		FilesID:     em.FilesIDs,
	}

	if err := s.storage.CreateExtraMessageForTicket(dbModel); err != nil {
		return err
	}

	dbLogModel := db.TicketLogModel{
		TicketID:  em.TicketID,
		ChangedBy: senderID,
		Action:    "Создано дополнительное сообщение",
	}

	err := s.storage.LogTicketOperation(dbLogModel)
	if err != nil {
		return err
	}

	return nil
}
