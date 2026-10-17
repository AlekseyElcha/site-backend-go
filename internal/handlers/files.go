package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/exceptions"
	"site-backend-go/internal/middleware"
	"uuid"
)

func (s *Server) DownloadFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dtos.DownloadFileRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed"}`))
		return
	}
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	fileIDs := []uuid.UUID{req.FileID}
	fileInfo, err := s.FileService.GetMultipleFilesInfo(ctx, fileIDs)
	if err != nil {
		if errors.Is(err, exceptions.ErrUnauthorized) {
			s.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(fileInfo) == 0 {
		s.WriteError(w, http.StatusNotFound, "file not found")
		return
	}

	fileName, err := s.FileService.GetFilenameByID(req.FileID)
	baseDir, err := filepath.Abs("./files")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filePath := filepath.Join(baseDir, fileName)

	http.ServeFile(w, r, filePath)
}

func (s *Server) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	uploaderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed"}`))
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.WriteError(w, http.StatusBadRequest, "Error parsing multipart form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		s.WriteError(w, http.StatusBadRequest, "Missing file parameter")
		return
	}
	defer file.Close()

	genFileID := uuid.New()
	originalFilename := header.Filename

	fileID, err := s.FileService.SaveFileLocally(file, genFileID, originalFilename)
	if err != nil {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		s.Logger.Error("error while saving file in storage", "fileID", fileID, "error", err)
		return
	}

	fileID_UUID, err := uuid.Parse(fileID)
	if err != nil {
		s.WriteError(w, http.StatusBadRequest, "Invalid file ID")
		return
	}

	err = s.FileService.SaveFileInfo(fileID_UUID, uploaderID, originalFilename)
	if err != nil {
		s.WriteError(w, http.StatusInternalServerError, "internal server error")
		s.Logger.Error("error while saving file info", "fileID", "uploaderID", uploaderID, "originalFilename", originalFilename, "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"file_id": fileID}
	json.NewEncoder(w).Encode(response)
}
