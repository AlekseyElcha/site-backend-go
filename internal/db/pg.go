package db

import (
	"database/sql"
	"fmt"
	"time"
	"uuid"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(database *sql.DB) *Storage {
	return &Storage{
		db: database,
	}
}

type UserCreateModel struct {
	Email string
}

type UserUpdateModel struct {
	Name        string
	Address     string
	PhoneNumber string
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
	Status string
}

type AnswerTicketModel struct {
	ID         uuid.UUID
	AnswerText string
}

type TicketInfo struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	Name        string
	Email       string
	PhoneNumber string
	Address     string
	Question    string
	Status      string
}

func InitDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS users (
	    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	    name TEXT DEFAULT NULL,
	    address TEXT DEFAULT NULL,
	    phone_number TEXT DEFAULT NULL,
	    role TEXT NOT NULL DEFAULT 'user',
	    email TEXT NOT NULL UNIQUE,
	    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS tickets (
		id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
	    name TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		email TEXT NOT NULL,
	    address TEXT NOT NULL,
		phone_number TEXT DEFAULT NULL,
		question TEXT NOT NULL,
	    files UUID[] DEFAULT NULL,
		status TEXT NOT NULL DEFAULT 'active',
	    CONSTRAINT fk_user
	    	FOREIGN KEY (user_id)
	    	REFERENCES users (id)
	); 
	
	CREATE TABLE IF NOT EXISTS answers (
		id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	    sender_id UUID NOT NULL,
		ticket_id UUID NOT NULL,
		answer_text TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_ticket 
			FOREIGN KEY (ticket_id) 
			REFERENCES tickets(id) 
	);

	CREATE TABLE IF NOT EXISTS files (
	    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	    uploader_id UUID NOT NULL,
	    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)
`

	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func (s *Storage) CreateUser(model UserCreateModel) error {
	query := `
		INSERT INTO users (email)
		VALUES ($1)
	`
	if _, err := s.db.Exec(query, model.Email); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetTicketInfoByID(id uuid.UUID) (*TicketInfo, error) {
	query := `
		SELECT id, created_at, question, status FROM tickets
		WHERE id = $1;
	`
	var t TicketInfo
	err := s.db.QueryRow(query, id).Scan(&t.ID, &t.CreatedAt, &t.Question, &t.Status)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &t, nil
}

func (s *Storage) CreateTicket(model CreateTicketModel) error {
	query := `
		INSERT INTO tickets (user_id, name, email, phone_number, address, question)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(
		query,
		model.UserID,
		model.Name,
		model.Email,
		model.PhoneNumber,
		model.Address,
		model.Question,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateTicketStatus(model UpdateTicketStatusModel) error {
	query := `
		UPDATE tickets
		SET status = $2
		WHERE id = $1`

	_, err := s.db.Exec(query, model.ID, model.Status)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) AnswerTicket(model AnswerTicketModel, senderID uuid.UUID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO answers (sender_id, ticket_id, answer_text)
		VALUES ($1, $2, $3)`

	if _, err := tx.Exec(insertQuery, senderID, model.ID, model.AnswerText); err != nil {
		return err
	}

	updateQuery := `
		UPDATE tickets
		SET status = 'answered'
		WHERE id = $1`

	if _, err := tx.Exec(updateQuery, model.ID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) AddFileInfoToDB(uploaderID uuid.UUID) error {
	query := `
		INSERT INTO files (uploader_id)
		VALUES ($1)
	`

	if _, err := s.db.Exec(query, uploaderID); err != nil {
		return err
	}

	return nil
}
