package controllers

import (
	"encoding/json"
	"file-sync/pkg/enums"
	globalmodels "file-sync/pkg/models"
	"io"
	"net/http"
	"server/models"
	"server/services"
)

type FileController struct {
	fileService services.FileService
}

func NewFileController(fileService services.FileService) *FileController {
	return &FileController{
		fileService,
	}
}

func (c *FileController) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		responseError(w, err, http.StatusInternalServerError)
		return
	}

	fileInfo := &globalmodels.FileInfo{}
	err = json.Unmarshal(body, &globalmodels.FileInfo{})
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}

	var status enums.FileStatus
	status, err = c.fileService.GetStatus(hash, fileInfo)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOk(w, status)
}

func (c *FileController) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	fileStream, err := c.fileService.GetFileStream(hash)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseStream(w, fileStream)
}

func (c *FileController) GetUploadSessionHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	sessionId, err := c.fileService.GetUploadSession(hash)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOk(w, sessionId)
}

func (c *FileController) UploadChunkHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	sessionId := r.PathValue("sessionId")
	chunk, err := io.ReadAll(r.Body)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	err = c.fileService.UploadChunk(hash, sessionId, chunk)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOk(w, nil)
}

func (c *FileController) CommitChunksHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	sessionId := r.PathValue("sessionId")
	err := c.fileService.CommitChunks(hash, sessionId)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	responseOk(w, nil)
}

func (c *FileController) GetSyncedFileMapHandler(w http.ResponseWriter, _ *http.Request) {
	fileMap := c.fileService.GetFileMap()
	responseOk(w, fileMap)
}

func responseOk(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if data == nil {
		response, err := json.Marshal(models.Response{Status: "success"})
		if err != nil {
			responseError(w, err, http.StatusInternalServerError)
			return
		}
		w.Write(response)
		return
	}
	response, err := json.Marshal(models.ResponseSuccess{
		Response: models.Response{Status: "success"},
		Data:     data,
	})
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func responseError(w http.ResponseWriter, error error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response, err := json.Marshal(models.ResponseError{
		Response: models.Response{Status: "error"},
		Error:    error.Error(),
	})
	if err != nil {
		w.Write([]byte(`{"status":"error","error":"Internal Server Error"}`))
	}
	w.Write(response)
}

func responseStream(w http.ResponseWriter, stream []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(stream)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
	}
}
