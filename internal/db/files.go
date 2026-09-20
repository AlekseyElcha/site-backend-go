package db

import (
	"fmt"
	"uuid"
)

func (s *Storage) AddFileInfoToDB(fileID uuid.UUID, uploaderID uuid.UUID) error {
	query := `
		INSERT INTO files (id, uploader_id)
		VALUES ($1, $2)
	`
	_, err := s.db.Exec(query, fileID, uploaderID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetFilenameByFileID(fileID uuid.UUID) (string, error) {
	query := `
		SELECT
			extension
		FROM files
		WHERE id = $1
	`
	fileExt, err := s.db.Exec(query, fileID)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s.%s", fileID.String(), fileExt)
	return fileName, nil
}
