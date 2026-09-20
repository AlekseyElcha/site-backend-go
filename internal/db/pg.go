package db

import (
	"database/sql"
	"time"

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
	    id UUID NOT NULL PRIMARY KEY,
	    extension TEXT DEFAULT NULL,
	    uploader_id UUID NOT NULL,
	    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	
	CREATE TABLE IF NOT EXISTS extra_messages (
	    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	    sender_id UUID NOT NULL,
	    ticket_id UUID NOT NULL,
	    message_text TEXT NOT NULL,
	    files UUID[] DEFAULT NULL,
	    CONSTRAINT fk_extra_messages_ticket
	    	FOREIGN KEY (ticket_id)
	    	REFERENCES tickets(id)
	);

	CREATE TABLE IF NOT EXISTS ticket_logs (
	    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
	    ticket_id UUID NOT NULL,
	    changed_by UUID NOT NULL,
	    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	    action TEXT NOT NULL,
	    CONSTRAINT fk_ticket
	    	FOREIGN KEY (ticket_id)
	    	REFERENCES tickets(id),
	    CONSTRAINT fk_users
			FOREIGN KEY (changed_by)
			REFERENCES users(id)
	);
`

	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
