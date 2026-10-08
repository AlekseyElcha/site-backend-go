package handlers

import (
	"encoding/json"
	"net/http"
	"site-backend-go/internal/dtos"
	"site-backend-go/internal/middleware"
	"uuid"
)

//type FileHandler struct {
//	FileService *service.FileService
//}
//
//func NewFileHandler(fs *service.FileService) *FileHandler {
//	return &FileHandler{FileService: fs}
//}

func (s *Server) DownloadFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	var req dtos.DownloadFileRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed"}`))
	}
	if err := validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed", "details":"` + err.Error() + `"}`))
		return
	}

	fileName, err := s.FileService.GetFilenameByID(req.FileID)

	filePath := "./files/" + fileName
	http.ServeFile(w, r, filePath)
}

func (s *Server) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	uploaderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Validation failed"}`))
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, "Error parsing multipart form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Missing file parameter")
		return
	}
	defer file.Close()

	genFileID := uuid.New()

	fileID, err := s.FileService.SaveFileLocally(file, genFileID, header.Filename)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		s.Logger.Error("error while saving file in storage", "fileID", fileID, "error", err)
		return
	}

	err = s.FileService.SaveFileInfo(uploaderID, genFileID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		s.Logger.Error("error while saving file info", "fileID", "genFileID", genFileID, fileID, "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"file_id": fileID}
	json.NewEncoder(w).Encode(response)
}
