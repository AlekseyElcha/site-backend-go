package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"site-backend-go/internal/exceptions"
	"uuid"
)

func (s *Storage) GetAllTicketsInfo(ctx context.Context) ([]TicketGeneralInfoModel, error) {
	query := `
		SELECT
			id
			, user_id
			, name
			, created_at
			, email
			, address
			, status
		FROM tickets
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Tickets []TicketGeneralInfoModel

	for rows.Next() {
		var t TicketGeneralInfoModel
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.Email,
			&t.Address,
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

func (s *Storage) GetTicketInfoByID(ctx context.Context, ID uuid.UUID) (*TicketInfoModel, error) {
	queryTicket := `
		SELECT 
		    t.id
		    , t.user_id
		    , t.name
			, t.created_at
		    , t.email
		    , t.address
		    , t.phone_number
		    , t.question
		    , t.files
		    , t.status
		FROM tickets t
		WHERE t.id = $1;
	`
	var t TicketInfoModel

	err := s.db.QueryRow(queryTicket, ID).Scan(
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

	return &t, nil
}

func (s *Storage) GetAllTicketsByUserID(ctx context.Context, userID uuid.UUID) ([]TicketGeneralInfoModel, error) {
	query := `
		SELECT
			id
			, user_id
			, name
			, created_at
			, email
			, address
-- 			, phone_number
-- 			, question
-- 			, files
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

	var tickets []TicketGeneralInfoModel

	for rows.Next() {
		var t TicketGeneralInfoModel
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.Email,
			&t.Address,
			&t.Status,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (s *Storage) GetAnswerForTicketByID(ctx context.Context, id uuid.UUID) (*TicketAnswersModel, error) { // TODO: нейминг
	query := `
		SELECT 
		    id
		    , sender_id
		    , ticket_id
		    , answer_text
		 	, files
		    , created_at
		FROM answers
		WHERE ticket_id = $1;
	`
	var ans TicketAnswersModel

	err := s.db.QueryRow(
		query, id,
	).Scan(
		&ans.ID,
		&ans.SenderID,
		&ans.TicketID,
		&ans.AnswerText,
		&ans.Files,
		&ans.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ans, err
}

func (s *Storage) GetExtraMessagesForTicketByID(ctx context.Context, id uuid.UUID) ([]ExtraMessageInfoModel, error) { // TODO: пересмотреть
	query := `
		SELECT 
		m.id,
		m.sender_id,
		m.ticket_id,
		m.message_text,
		u.role AS sender_role,
		COALESCE(
			json_agg(
				json_build_object(
					'file_id', f.id,
					'uploader_id', f.uploader_id,
					'original_filename', f.original_filename,
					'uploaded_at', f.uploaded_at
				)
			) FILTER (WHERE f.id IS NOT NULL), 
			'[]'
   		) AS files
	FROM extra_messages m
	JOIN users u ON m.sender_id = u.id
	LEFT JOIN files f ON f.id = ANY(m.files)
	WHERE m.ticket_id = $1
	GROUP BY m.id, u.role;		
	`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extraMessages []ExtraMessageInfoModel

	for rows.Next() {
		var msg ExtraMessageInfoModel
		var filesJSON []byte

		err = rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.TicketID,
			&msg.MessageText,
			&msg.SenderRole,
			&filesJSON,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(filesJSON, &msg.Files); err != nil {
			return nil, err
		}

		extraMessages = append(extraMessages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(extraMessages) == 0 {
		return nil, exceptions.ErrNotFound
	}

	return extraMessages, nil
}

func (s *Storage) CreateTicket(ctx context.Context, model CreateTicketModel) (uuid.UUID, error) {
	query := `
		INSERT INTO tickets (
		                     user_id
		                     , name
		                     , email
		                     , phone_number
		                     , address
		                     , files
		                     , question
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
		model.FilesIDs,
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

func (s *Storage) UpdateTicketStatus(ctx context.Context, model UpdateTicketStatusModel) error {
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

func (s *Storage) AnswerTicket(ctx context.Context, model AnswerTicketModel, senderID uuid.UUID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := `
		INSERT INTO answers (sender_id, ticket_id, answer_text, files)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(insertQuery, senderID, model.ID, model.AnswerText, model.FilesIDs)
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

func (s *Storage) CreateExtraMessageForTicket(ctx context.Context, model ExtraMessageModel) error {
	query := `
		INSERT INTO extra_messages(ticket_id, sender_id, message_text, files)
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

func (s *Storage) LogTicketOperation(ctx context.Context, model TicketLogModel) error {
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

func (s *Storage) SearchForTickets(ctx context.Context, userID string, userRole string, searchingFor string) ([]TicketGeneralInfoModel, error) {
	query := `
		SELECT
		    id
			, user_id
			, name
			, created_at
			, email
			, address
			, status
		FROM tickets
		WHERE
		    ($2 = 'admin' AND (
		        name ILIKE $3
				OR email ILIKE $3
				OR phone_number ILIKE $3
				OR address ILIKE $3
				OR question ILIKE $3
		    ))
		    OR
			($2 != 'admin' AND user_id = $1)
		ORDER BY created_at DESC
	`

	searchingForFormatted := "%" + searchingFor + "%"
	rows, err := s.db.QueryContext(ctx, query, userID, userRole, searchingForFormatted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suitableTickets := make([]TicketGeneralInfoModel, 0)
	for rows.Next() {
		var ticketInfo TicketGeneralInfoModel
		err := rows.Scan(
			&ticketInfo.ID,
			&ticketInfo.UserID,
			&ticketInfo.Name,
			&ticketInfo.CreatedAt,
			&ticketInfo.Email,
			&ticketInfo.Address,
			&ticketInfo.Status,
		)
		if err != nil {
			s.log.Error("Error while Scanning suitable ticket", "err", err)
			return nil, exceptions.ErrDBError
		}
		suitableTickets = append(suitableTickets, ticketInfo)
	}

	if err := rows.Err(); err != nil {
		s.log.Error("rows error after iteration", "err", err)
		return nil, exceptions.ErrDBError
	}

	return suitableTickets, err
}
