package service

import (
	"io"
	"os"
	"path/filepath"
	"uuid"
)

type FilesDB interface {
	AddFileInfoToDB(uploaderID uuid.UUID) error
}

type FileService struct {
	uploadDir string
	storage   FilesDB
}

func NewFileService(uploadDir string) *FileService {
	return &FileService{
		uploadDir: uploadDir,
	}
}

func (s *FileService) SaveFileLocally(
	fileData io.Reader,
	originalFilename string,
) (string, error) {
	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	fileID := uuid.New().String()
	ext := filepath.Ext(originalFilename)
	uniqueFilename := fileID + ext

	dstPath := filepath.Join(s.uploadDir, uniqueFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, fileData); err != nil {
		return "", err
	}

	return fileID, nil
}

func (s *FileService) SaveFileInfo(uploaderID uuid.UUID) error {
	return s.storage.AddFileInfoToDB(uploaderID)
}
