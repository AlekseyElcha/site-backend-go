package db

import (
	"context"
	"path/filepath"
	"uuid"
)

func (s *Storage) AddFileInfoToDB(fileID uuid.UUID, originalFilename string, uploaderID uuid.UUID) error {
	query := `
		INSERT INTO files (id, uploader_id, original_filename)
		VALUES ($1, $2, $3)
	`
	_, err := s.db.Exec(query, fileID, uploaderID, originalFilename)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetFileInfoByFileID(fileID uuid.UUID) (string, error) {
	query := `
		SELECT original_filename
		FROM files
		WHERE id = $1
	`
	var originalFilename string
	err := s.db.QueryRow(query, fileID).Scan(&originalFilename)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(originalFilename)
	fileName := fileID.String() + ext
	return fileName, nil
}

func (s *Storage) GetFilesInfoByFilesIDMultiple(ctx context.Context, fileIDs []uuid.UUID) ([]FileInfoModel, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT
		    id
		    , uploader_id
			, original_filename
			, uploaded_at
		FROM files
		WHERE id = ANY($1::uuid[])
	`

	rows, err := s.db.QueryContext(ctx, query, fileIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := make([]FileInfoModel, 0, len(fileIDs))

	for rows.Next() {
		var item FileInfoModel
		err := rows.Scan(
			&item.FileID,
			&item.UploaderID,
			&item.OriginalFilename,
			&item.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		info = append(info, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return info, nil
}
