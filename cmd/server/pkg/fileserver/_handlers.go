package fileserver

import (
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	log "github.com/sirupsen/logrus"
	"io"
)

func (m *Mux) handleSyncRequest(transactionID string, request globalmodels.SyncRequest) {
	responseMessage := globalmodels.Message{
		TransactionID: transactionID,
	}

	transactionError := false
	fileInfo := request.Info
	syncedFileInfo, found := m.fileService.GetFileInfo(request.Hash)
	if !found {
		fileInfo.Status = globalenums.New
	} else if fileInfo.Checksum == syncedFileInfo.Checksum {
		fileInfo.Status = globalenums.Synced
	} else if fileInfo.ModTime().After(syncedFileInfo.ModTime()) {
		fileInfo.Status = globalenums.Dirty
	} else if syncedFileInfo.ModTime().After(fileInfo.ModTime()) {
		fileInfo.Status = globalenums.Stale
	} else {
		transactionError = true
	}

	if transactionError {
		responseMessage.Payload = globalmodels.Error{
			Message: "Error processing sync request",
		}
	} else {
		responseMessage.Payload = globalmodels.SyncResponse{
			Info: fileInfo,
		}
	}
	m.messageChan <- responseMessage
}

// handleDownloadRequest handles a download request from the client.
func (m *Mux) handleDownloadRequest(transactionID string, request globalmodels.DownloadRequest) {
	responseMessage := globalmodels.Message{
		TransactionID: transactionID,
	}

	fileBuffer, err := m.fileService.GetFile(request.Hash)
	if err != nil {
		log.Error("Error getting filestream: ", err)
		responseMessage.Payload = globalmodels.Error{
			Message: "Error getting filestream",
		}
		m.messageChan <- responseMessage
		return
	}

	var n int
	done := false
	for {
		chunk := make([]byte, m.chunkSize)
		n, err = fileBuffer.Read(chunk)
		if err != nil && err != io.EOF {
			log.Error("Error reading filestream: ", err)
			responseMessage.Payload = globalmodels.Error{
				Message: "Error reading filestream",
			}
			m.messageChan <- responseMessage
			return
		} else if err == io.EOF || n == 0 {
			done = true
		}
		responseMessage.Payload = globalmodels.FileChunk{
			Chunk: chunk[:n],
			Done:  done,
		}
		m.messageChan <- responseMessage
		if done {
			break
		}
	}

}

// handleUploadRequest handles an upload request from the client.
func (m *Mux) handleUploadRequest(transactionID string, request globalmodels.UploadRequest) {
	m.uploadChannels[transactionID] = make(chan globalmodels.FileChunk, 5)

	fileBuffer := make([]byte, 0)
	var chunk globalmodels.FileChunk
	for {
		chunk = <-m.uploadChannels[transactionID]

		fileBuffer = append(fileBuffer, chunk.Chunk...)
		if chunk.Done {
			break
		}
	}

	err := m.fileService.CreateFile(request.Hash, request.Checksum, fileBuffer)
	responseMessage := globalmodels.Message{
		TransactionID: transactionID,
	}
	if err != nil {
		log.Error("Error creating file: ", err)
		responseMessage.Payload = globalmodels.UploadResult{
			Hash:    request.Hash,
			Success: false,
		}
	} else {
		responseMessage.Payload = globalmodels.UploadResult{
			Hash:    request.Hash,
			Success: true,
		}
	}
	delete(m.uploadChannels, transactionID)
	m.messageChan <- responseMessage
}

// handleFileChunk handles a file chunk from the client.
func (m *Mux) handleFileChunk(transactionID string, request globalmodels.FileChunk) {
	uploadChannel, found := m.uploadChannels[transactionID]
	if !found {
		errorMessage := globalmodels.Message{
			TransactionID: transactionID,
			Payload: globalmodels.Error{
				Message: "Upload channel not found",
			},
		}
		m.messageChan <- errorMessage
		return
	}
	uploadChannel <- request
}
