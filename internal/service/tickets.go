package service

import (
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"uuid"
)

type TicketStorage interface {
	GetTicketInfoByID(id uuid.UUID) (*db.TicketInfo, error)
	CreateTicket(model db.CreateTicketModel) error
	UpdateTicketStatus(model db.UpdateTicketStatusModel) error
	AnswerTicket(model db.AnswerTicketModel, senderID uuid.UUID) error
}
type TicketService struct {
	storage TicketStorage
}

func NewTicketService(storage TicketStorage) *TicketService {
	return &TicketService{
		storage: storage,
	}
}

func (s *TicketService) GetTicket(id uuid.UUID) (*db.TicketInfo, error) {
	t, err := s.storage.GetTicketInfoByID(id)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *TicketService) Create(t dtos.TicketCreateRequest, userID uuid.UUID) error {
	dbModel := db.CreateTicketModel{
		UserID:      userID,
		Name:        t.Name,
		Email:       t.Email,
		PhoneNumber: t.PhoneNumber,
		Address:     t.Address,
		Question:    t.Question,
	}

	if err := s.storage.CreateTicket(dbModel); err != nil {
		return err
	}

	return nil
}

func (s *TicketService) UpdateStatus(t dtos.TicketStatusUpdateRequest) error {
	dbModel := db.UpdateTicketStatusModel{
		ID:     t.ID,
		Status: t.Status,
	}

	if err := s.storage.UpdateTicketStatus(dbModel); err != nil {
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

	return nil
}
