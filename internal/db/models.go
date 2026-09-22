package db

import (
	"time"
	"uuid"
)

type UserInfoModel struct {
	ID          uuid.UUID
	Email       *string
	PhoneNumber *string
	Name        *string
	Address     *string
	Role        *string
}

type UserCreateModel struct {
	Email string
}

type UserUpdateModel struct {
	ID          uuid.UUID `db:"id"`
	Name        *string   `db:"name"`
	Address     *string   `db:"address"`
	PhoneNumber *string   `db:"phone_number"`
}

type CreateTicketModel struct {
	UserID      uuid.UUID
	Name        string
	Email       string
	PhoneNumber string
	Address     string
	Question    string
}

type UpdateTicketStatusModel struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Status string
}

type AnswerTicketModel struct {
	ID         uuid.UUID
	AnswerText string
}

type TicketAnswersInfo struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	MessageText string
}

type TicketInfoModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CreatedAt   time.Time
	Name        string
	Email       string
	PhoneNumber string
	Address     string
	Question    string
	Files       []uuid.UUID
	Status      string
	Answer      TicketAnswersInfo
}

type TicketGeneralInfoModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CreatedAt   time.Time
	Name        string
	Email       string
	PhoneNumber string
	Address     string
	Question    string
	Files       []uuid.UUID
	Status      string
}

type TicketAnswersModel struct {
	ID         uuid.UUID `db:"id"`
	SenderID   uuid.UUID `db:"sender_id"`
	TicketID   uuid.UUID `db:"ticket_id"`
	AnswerText string    `db:"answer_text"`
	CreatedAt  time.Time `db:"created_at"`
}

type TicketExtraMessagesModel struct {
	ID          uuid.UUID   `db:"id"`
	SenderID    uuid.UUID   `db:"sender_id"`
	TicketID    uuid.UUID   `db:"ticket_id"`
	MessageText string      `db:"message_text"`
	Files       []uuid.UUID `db:"files"`
}
type ExtraMessageModel struct {
	TicketID    uuid.UUID
	SenderID    uuid.UUID
	MessageText string
	FilesID     []uuid.UUID
}

type ExtraMessageInfoModel struct {
	ID          uuid.UUID   `db:"id"`
	TicketID    uuid.UUID   `db:"ticket_id"`
	SenderID    uuid.UUID   `db:"sender_id"`
	SenderRole  string      `db:"sender_role"`
	MessageText string      `db:"message_text"`
	FilesID     []uuid.UUID `db:"files_id"`
}

type TicketLogModel struct {
	TicketID  uuid.UUID
	ChangedBy uuid.UUID
	Action    string
}
