package handlers

import (
	"encoding/json"
	"net/http"
	"site-backend-go/internal/auth"
	"site-backend-go/internal/service"
)

type FileHandler struct {
	FileService *service.FileService
}

func NewFileHandler(fs *service.FileService) *FileHandler {
	return &FileHandler{FileService: fs}
}

func DownloadFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	filePath := "/files/test.text"
	http.ServeFile(w, r, filePath)
}

func (f *FileHandler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
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

	fileID, err := f.FileService.SaveFileLocally(file, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = f.FileService.SaveFileInfo(uploaderID)
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
