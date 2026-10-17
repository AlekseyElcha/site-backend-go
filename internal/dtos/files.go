package dtos

import "uuid"

type DownloadFileRequest struct {
	FileID uuid.UUID `json:"file_id"`
}

type FileDetailsInfo struct {
	FileID           uuid.UUID `json:"file_id"`
	UploaderID       uuid.UUID `json:"uploader_id"`
	OriginalFilename string    `json:"original_filename"`
	UploadedAt       string    `json:"uploaded_at"`
}
