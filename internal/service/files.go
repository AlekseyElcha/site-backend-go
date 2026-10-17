package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"site-backend-go/internal/db"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"uuid"
)

type FilesDB interface {
	AddFileInfoToDB(fileID uuid.UUID, originalFilename string, uploaderID uuid.UUID) error
	GetFileInfoByFileID(fileID uuid.UUID) (string, error)
	GetFilesInfoByFilesIDMultiple(ctx context.Context, fileIDs []uuid.UUID) ([]db.FileInfoModel, error)
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

func (s *FileService) SaveFileInfo(fileID uuid.UUID, uploaderID uuid.UUID, originalFilename string) error {
	err := s.storage.AddFileInfoToDB(fileID, originalFilename, uploaderID)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) GetMultipleFilesInfo(ctx context.Context, fileIDs []uuid.UUID) ([]dtos.FileDetailsInfo, error) {
	filesInfo, err := s.storage.GetFilesInfoByFilesIDMultiple(ctx, fileIDs)
	if err != nil {
		return nil, err
	}

	filesInfoDto := make([]dtos.FileDetailsInfo, 0, len(filesInfo))
	for _, fileInfo := range filesInfo {
		info := dtos.FileDetailsInfo{
			FileID:           fileInfo.FileID,
			UploaderID:       fileInfo.UploaderID,
			OriginalFilename: fileInfo.OriginalFilename,
			UploadedAt:       fileInfo.UploadedAt,
		}

		filesInfoDto = append(filesInfoDto, info)
	}

	return filesInfoDto, nil
}

func (s *FileService) GetFilenameByID(fileID uuid.UUID) (string, error) {
	fileName, err := s.storage.GetFileInfoByFileID(fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", exceptions.ErrNotFound
		}
		return "", err
	}
	return fileName, nil
}
