package dtos

import "uuid"

type ExtraMessageCreateRequest struct {
	TicketID    uuid.UUID   `json:"ticket_id" validate:"required"`
	MessageText string      `json:"message_text" validate:"required"`
	FilesIDs    []uuid.UUID `json:"files_ids" validate:"omitempty"`
}
