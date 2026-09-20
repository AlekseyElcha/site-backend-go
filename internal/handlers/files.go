package handlers

import (
	"encoding/json"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/dtos"
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
	uploaderID, err := auth.GetUserIDFromCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}
	defer file.Close()

	genFileID := uuid.New()

	fileID, err := s.FileService.SaveFileLocally(file, genFileID, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.FileService.SaveFileInfo(uploaderID, genFileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"file_id": fileID}
	json.NewEncoder(w).Encode(response)
}
