package dtos

import (
	"time"
	"uuid"
)

type TicketGetResponse struct {
	ID          uuid.UUID                   `json:"id"`
	UserID      uuid.UUID                   `json:"user_id"`
	Name        string                      `json:"name"`
	Address     string                      `json:"address"`
	PhoneNumber string                      `json:"phone_number"`
	CreatedAt   time.Time                   `json:"created_at"`
	Question    string                      `json:"question"`
	Status      string                      `json:"status"`
	FilesIDs    []uuid.UUID                 `json:"files_ids"`
	Answers     []TicketAnswersInfoResponse `json:"answers"`
}

type TicketAnswersInfoResponse struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	AnswerText string    `json:"answer_text"`
}

type TicketCreateRequest struct {
	Name        string      `json:"name" validate:"required"`
	Email       string      `json:"email" validate:"required,email"`
	PhoneNumber string      `json:"phone_number" validate:"required"`
	Address     string      `json:"address" validate:"required"`
	Question    string      `json:"question" validate:"required"`
	FilesIDs    []uuid.UUID `json:"files_ids" validate:"omitempty"`
}

type TicketStatusUpdateRequest struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	Status string    `json:"status" validate:"required"`
}

type TicketAnswerRequest struct {
	TicketID   uuid.UUID `json:"ticket_id" validate:"required"`
	AnswerText string    `json:"answer_text" validate:"required"`
}

type ExtraMessagesGetRequest struct {
	TicketID uuid.UUID `json:"ticket_id" validate:"required"`
}
