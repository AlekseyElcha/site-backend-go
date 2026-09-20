package service

import (
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"site-backend-go/internal/exceptions"
	"uuid"
)

type FilesDB interface {
	AddFileInfoToDB(fileID uuid.UUID, uploaderID uuid.UUID) error
	GetFilenameByFileID(fileID uuid.UUID) (string, error)
}

type FileService struct {
	uploadDir string
	storage   FilesDB
}

func NewFileService(uploadDir string, storage FilesDB) *FileService {
	return &FileService{
		uploadDir: uploadDir,
		storage:   storage,
	}
}

func (s *FileService) SaveFileLocally(
	fileData io.Reader,
	fileID uuid.UUID,
	originalFilename string,
) (string, error) {
	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		return "", err
	}
	//
	//fileID := uuid.New().String()
	ext := filepath.Ext(originalFilename)
	uniqueFilename := fileID.String() + ext

	dstPath := filepath.Join(s.uploadDir, uniqueFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, fileData); err != nil {
		return "", err
	}

	return fileID.String(), nil
}

func (s *FileService) SaveFileInfo(uploaderID uuid.UUID, fileID uuid.UUID) error {
	err := s.storage.AddFileInfoToDB(fileID, uploaderID)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) DownloadFileByID(fileId uuid.UUID) {
	//
}

func (s *FileService) GetFilenameByID(fileID uuid.UUID) (string, error) {
	fileName, err := s.storage.GetFilenameByFileID(fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", exceptions.ErrNotFound
		}
		return "", err
	}

	_, err = os.Stat("./files/" + fileName)
	if err == nil {
		return fileName, nil
	} else {
		return "", exceptions.ErrNotFound
	}
}
