package dtos

import "uuid"

type DownloadFileRequest struct {
	FileID uuid.UUID `json:"file_id"`
}
