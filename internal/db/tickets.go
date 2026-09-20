package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"site-backend-go/internal/exceptions"
	"uuid"
)

func (s *Storage) GetAllTicketsInfo() ([]TicketInfoModel, error) {
	query := `
		SELECT
			id
			, user_id
			, name
			, created_at
			, email
			, address
			, phone_number
			, question
			, files
			, status
		FROM tickets
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Tickets []TicketInfoModel

	for rows.Next() {
		var t TicketInfoModel
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.Email,
			&t.Address,
			&t.PhoneNumber,
			&t.Question,
			&t.Files,
			&t.Status,
		)
		if err != nil {
			return nil, err
		}

		Tickets = append(Tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return Tickets, nil
}

func (s *Storage) GetTicketInfoByID(id uuid.UUID) (*TicketInfoModel, error) {
	queryTicket := `
		SELECT 
		    t.id
		    , t.user_id
			, t.created_at
		    , t.question
		    , t.status
		    , a.answer_text 
		FROM tickets t
		JOIN answers a ON t.id = a.ticket_id
		WHERE t.id = $1;
	`
	var t TicketInfoModel // TODO: здесь
	err := s.db.QueryRow(queryTicket, id).Scan(&t.ID, &t.CreatedAt, &t.Question, &t.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		fmt.Println(err)
		return nil, err
	}

	return &t, nil
}

func (s *Storage) GetAllTicketsByUserID(userID uuid.UUID) ([]TicketInfoModel, error) {
	query := `
		SELECT
			id
			, user_id
			, name
			, created_at
			, email
			, address
			, phone_number
			, question
			, files
			, status
		FROM tickets
		WHERE user_id = $1;
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	var tickets []TicketInfoModel

	for rows.Next() {
		var t TicketInfoModel
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.Email,
			&t.Address,
			&t.PhoneNumber,
			&t.Question,
			&t.Files,
			&t.Status,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (s *Storage) GetAnswersForTicketByID(id uuid.UUID) ([]TicketAnswersModel, error) {
	query := `
		SELECT 
		    id
		    , sender_id
		    , ticket_id
		    , answer_text
		    , created_at
		FROM tickets
		WHERE id = $1;
	`

	//var t TicketAnswersModel
	//err := s.db.QueryRow(query, id).Scan(&t.ID, &t.SenderID, &t.TicketID, &t.AnswerText, &t.CreatedAt)
	//if err != nil {
	//	if errors.Is(err, sql.ErrNoRows) {
	//		return nil, exceptions.ErrNotFound
	//	}
	//	fmt.Println(err)
	//	return nil, err
	//}

	rows, err := s.db.Query(query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		return nil, err
	}

	var answers []TicketAnswersModel

	for rows.Next() {
		var t TicketAnswersModel
		err := rows.Scan(
			&t.ID,
			&t.SenderID,
			&t.TicketID,
			&t.AnswerText,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		answers = append(answers, t)
	}

	return answers, nil
}

func (s *Storage) GetExtraMessagesForTicketByID(id uuid.UUID) ([]ExtraMessageInfoModel, error) {
	query := `
		SELECT 
		    id
		 	, sender_id
		 	, ticket_id
		    , message_text
			, files
		FROM extra_messages
		WHERE id = $1;
	`

	rows, err := s.db.Query(query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	var extraMessages []ExtraMessageInfoModel

	for rows.Next() {
		var msg ExtraMessageInfoModel

		err = rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.TicketID,
			&msg.MessageText,
			&msg.FilesID,
		)
		if err != nil {
			return nil, err
		}

		extraMessages = append(extraMessages, msg)
	}

	return extraMessages, nil
}

func (s *Storage) CreateTicket(ctx context.Context, model CreateTicketModel) (uuid.UUID, error) {
	query := `
		INSERT INTO tickets (user_id, name, email, phone_number, address, question)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil(), err
	}

	defer tx.Rollback()

	var ticketID uuid.UUID
	err = tx.QueryRowContext(
		ctx,
		query,
		model.UserID,
		model.Name,
		model.Email,
		model.PhoneNumber,
		model.Address,
		model.Question,
	).Scan(&ticketID)
	if err != nil {
		return uuid.Nil(), err
	}

	err = tx.Commit()
	if err != nil {
		return uuid.Nil(), err
	}

	return ticketID, nil
}

func (s *Storage) UpdateTicketStatus(model UpdateTicketStatusModel) error {
	query := `
		UPDATE tickets
		SET status = $2
		WHERE id = $1`

	_, err := s.db.Exec(query, model.ID, model.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.ErrNotFound
		}
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

	_, err = tx.Exec(insertQuery, senderID, model.ID, model.AnswerText)
	if err != nil {
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

func (s *Storage) CreateExtraMessageForTicket(model ExtraMessageModel) error {
	query := `
		INSERT INTO extra_messages(ticket_id, sender_id, message_text, filesdco)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := s.db.Exec(
		query,
		model.TicketID,
		model.SenderID,
		model.MessageText,
		model.FilesID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Storage) LogTicketOperation(model TicketLogModel) error {
	query := `
		INSERT INTO ticket_logs(ticket_id, changed_by, action)
		VALUES ($1, $2, $3)
	`

	if _, err := s.db.Exec(
		query,
		model.TicketID,
		model.ChangedBy,
		model.Action,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.ErrNotFound
		}
		return err
	}

	return nil
}
