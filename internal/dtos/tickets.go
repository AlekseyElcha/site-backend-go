package dtos

import (
	"time"
	"uuid"
)

type ItemResponse[T any] struct {
	Items []T `json:"items"`
}

type TicketGetResponse struct {
	Info          *TicketDetailedInfoResponse           `json:"info"`
	Files         *ItemResponse[FileDetailsInfo]        `json:"files"`
	ExtraMessages *ItemResponse[TicketExtraMessageInfo] `json:"extra_messages"`
	Answers       *TicketAnswerInfoResponse             `json:"answers"`
}

type TicketDetailedInfoResponse struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Address     string      `json:"address"`
	PhoneNumber string      `json:"phone_number"`
	CreatedAt   time.Time   `json:"created_at"`
	Question    string      `json:"question"`
	Status      string      `json:"status"`
	FilesIDs    []uuid.UUID `json:"files_ids,omitempty"`
}

type TicketGeneralInfo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
}

type TicketAnswerInfoResponse struct { // TODO: пересмотреть
	ID          uuid.UUID   `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	MessageText string      `json:"answer_text"`
	FilesIDs    []uuid.UUID `json:"files_ids" validate:"omitempty"`
}

type TicketAttachedFileInfo struct {
	FileID           string    `json:"file_id"`
	UploaderID       string    `json:"uploader_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

type TicketExtraMessageInfo struct {
	ID          uuid.UUID                `json:"id"`
	TicketID    uuid.UUID                `json:"ticket_id"`
	SenderID    uuid.UUID                `json:"sender_id"`
	SenderRole  string                   `json:"sender_role"`
	MessageText string                   `json:"message_text"`
	Files       []TicketAttachedFileInfo `json:"files"`
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
	TicketID   uuid.UUID   `json:"ticket_id" validate:"required"`
	AnswerText string      `json:"answer_text" validate:"required"`
	FilesIDs   []uuid.UUID `json:"files_ids" validate:"omitempty"`
}

type ExtraMessagesGetRequest struct {
	TicketID uuid.UUID `json:"ticket_id" validate:"required"`
}

type StreetsAlikeResponse struct {
	Streets []string `json:"streets"`
}

type HousesAlikeResponse struct {
	Houses []string `json:"houses"`
}

type AddressesAlikeResponse struct {
	FullAddresses []string `json:"addresses"`
}

type TicketAlikeSearch struct {
	InputData string `json:"search_for"`
}
