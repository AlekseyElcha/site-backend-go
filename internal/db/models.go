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
	UserID      uuid.UUID   `db:"user_id"`
	Name        string      `db:"name"`
	Email       string      `db:"email"`
	PhoneNumber string      `db:"phone_number"`
	Address     string      `db:"address"`
	Question    string      `db:"question"`
	FilesIDs    []uuid.UUID `db:"files"`
}

type UpdateTicketStatusModel struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Status string
}

type AnswerTicketModel struct {
	ID         uuid.UUID
	AnswerText string
	FilesIDs   *[]uuid.UUID
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
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	Name      string
	Email     string
	Address   string
	Status    string
}

type TicketAnswersModel struct {
	ID         uuid.UUID   `db:"id"`
	SenderID   uuid.UUID   `db:"sender_id"`
	TicketID   uuid.UUID   `db:"ticket_id"`
	AnswerText string      `db:"answer_text"`
	Files      []uuid.UUID `db:"files"`
	CreatedAt  time.Time   `db:"created_at"`
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

type AttachedFileModel struct {
	FileID           string    `json:"file_id"`
	UploaderID       string    `json:"uploader_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

type ExtraMessageInfoModel struct {
	ID          uuid.UUID           `db:"id"`
	TicketID    uuid.UUID           `db:"ticket_id"`
	SenderID    uuid.UUID           `db:"sender_id"`
	SenderRole  string              `db:"sender_role"`
	MessageText string              `db:"message_text"`
	Files       []AttachedFileModel `db:"files"`
}

type TicketLogModel struct {
	TicketID  uuid.UUID
	ChangedBy uuid.UUID
	Action    string
}

type StreetAlikeModel struct {
	Name string `db:"street"`
}

type HouseAlikeModel struct {
	Number string `db:"house"`
}

type AddressAlikeModel struct {
	FullAddress string `db:"full_address"`
}

type FileInfoModel struct {
	FileID           uuid.UUID `db:"id"`
	UploaderID       uuid.UUID `db:"uploader_id"`
	OriginalFilename string    `db:"original_filename"`
	UploadedAt       string    `db:"uploaded_at"`
}
