package dtos

import (
	"time"
	"uuid"
)

type TicketGetResponse struct {
	ID            uuid.UUID                 `json:"id"`
	UserID        uuid.UUID                 `json:"user_id"`
	Name          string                    `json:"name"`
	Email         string                    `json:"email"`
	Address       string                    `json:"address"`
	PhoneNumber   string                    `json:"phone_number"`
	CreatedAt     time.Time                 `json:"created_at"`
	Question      string                    `json:"question"`
	Status        string                    `json:"status"`
	FilesIDs      []uuid.UUID               `json:"files_ids"`
	Answer        *TicketAnswerInfoResponse `json:"answer"`
	ExtraMessages *[]TicketExtraMessageInfo `json:"extra_messages"` // TODO: пересмотреть
}

type TicketGeneralInfo struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	CreatedAt   time.Time   `json:"created_at"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	PhoneNumber string      `json:"phone_number"`
	Address     string      `json:"address"`
	Question    string      `json:"question"`
	Files       []uuid.UUID `json:"files"`
	Status      string      `json:"status"`
}

type TicketAnswerInfoResponse struct { // TODO: пересмотреть
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	MessageText string    `json:"answer_text"`
}

type TicketExtraMessageInfo struct {
	ID          uuid.UUID   `json:"id"`
	TicketID    uuid.UUID   `json:"ticket_id"`
	SenderID    uuid.UUID   `json:"sender_id"`
	SenderRole  string      `json:"sender_role"`
	MessageText string      `json:"message_text"`
	FilesID     []uuid.UUID `json:"files_id"`
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
