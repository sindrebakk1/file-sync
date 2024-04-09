package fileserver

import (
	"encoding/gob"
	"file-sync/pkg/globalenums"
	"file-sync/pkg/globalmodels"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"server/pkg/auth"
	"server/services"
)

// TODO: Remake as generic MUX server with mux.ServeConn(net.Conn) and mux.Handle() methods

// clientHandler handles client connections.
type clientHandler struct {
	conn           net.Conn
	authenticator  auth.Authenticator
	fileService    services.FileService
	messageChan    chan globalmodels.Message
	chunkSize      int
	uploadChannels map[string]chan globalmodels.FileChunk
}

func newClientHandler(conn net.Conn, authenticator auth.Authenticator) *clientHandler {
	return &clientHandler{
		conn,
		authenticator,
		nil,
		make(chan globalmodels.Message, 5),
		1024,
		make(map[string]chan globalmodels.FileChunk),
	}
}

func (h *clientHandler) handleClient() {
	defer h.conn.Close()

	go func() {
		for message := range h.messageChan {
			err := h.sendMessage(&message)
			if err != nil {
				log.Error("Error sending message: ", err)
			}
		}
	}()

	// Initialize the client connection
	err := h.initClientConnection()
	if err != nil {
		log.Error("Error initializing client connection: ", err)
		return
	}

	// Handle client requests
	for {
		var request globalmodels.Message
		err = h.receiveMessage(&request)
		if err != nil {
			log.Error("Error decoding request: ", err)
			return
		}

		switch payload := request.Payload.(type) {
		case globalmodels.SyncRequest:
			go h.handleSyncRequest(request.TransactionID, payload)
		case globalmodels.DownloadRequest:
			go h.handleDownloadRequest(request.TransactionID, payload)
		case globalmodels.UploadRequest:
			go h.handleUploadRequest(request.TransactionID, payload)
		case globalmodels.FileChunk:
			go h.handleFileChunk(request.TransactionID, payload)
		case globalmodels.CloseConnection:
			return
		default:
			log.Error("Unknown request type: ", request)
		}
	}
}

func (h *clientHandler) handleSyncRequest(transactionID string, request globalmodels.SyncRequest) {
	responseMessage := globalmodels.Message{
		TransactionID: transactionID,
	}

	transactionError := false
	fileInfo := request.Info
	syncedFileInfo, found := h.fileService.GetFileInfo(request.Hash)
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
	h.messageChan <- responseMessage
}

// handleDownloadRequest handles a download request from the client.
func (h *clientHandler) handleDownloadRequest(transactionID string, request globalmodels.DownloadRequest) {
	responseMessage := globalmodels.Message{
		TransactionID: transactionID,
	}

	fileBuffer, err := h.fileService.GetFile(request.Hash)
	if err != nil {
		log.Error("Error getting filestream: ", err)
		responseMessage.Payload = globalmodels.Error{
			Message: "Error getting filestream",
		}
		h.messageChan <- responseMessage
		return
	}

	var n int
	done := false
	for {
		chunk := make([]byte, h.chunkSize)
		n, err = fileBuffer.Read(chunk)
		if err != nil && err != io.EOF {
			log.Error("Error reading filestream: ", err)
			responseMessage.Payload = globalmodels.Error{
				Message: "Error reading filestream",
			}
			h.messageChan <- responseMessage
			return
		} else if err == io.EOF || n == 0 {
			done = true
		}
		responseMessage.Payload = globalmodels.FileChunk{
			Chunk: chunk[:n],
			Done:  done,
		}
		h.messageChan <- responseMessage
		if done {
			break
		}
	}

}

// handleUploadRequest handles an upload request from the client.
func (h *clientHandler) handleUploadRequest(transactionID string, request globalmodels.UploadRequest) {
	h.uploadChannels[transactionID] = make(chan globalmodels.FileChunk, 5)

	fileBuffer := make([]byte, 0)
	var chunk globalmodels.FileChunk
	for {
		chunk = <-h.uploadChannels[transactionID]

		fileBuffer = append(fileBuffer, chunk.Chunk...)
		if chunk.Done {
			break
		}
	}

	err := h.fileService.CreateFile(request.Hash, request.Checksum, fileBuffer)
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
	delete(h.uploadChannels, transactionID)
	h.messageChan <- responseMessage
}

// handleFileChunk handles a file chunk from the client.
func (h *clientHandler) handleFileChunk(transactionID string, request globalmodels.FileChunk) {
	uploadChannel, found := h.uploadChannels[transactionID]
	if !found {
		errorMessage := globalmodels.Message{
			TransactionID: transactionID,
			Payload: globalmodels.Error{
				Message: "Upload channel not found",
			},
		}
		h.messageChan <- errorMessage
		return
	}
	uploadChannel <- request
}

// initClientConnection authenticates the client and initializes the file service for the current user.
func (h *clientHandler) initClientConnection() error {
	// Authenticate the client
	_, err := h.authenticator.AuthenticateClient()
	if err != nil {
		return err
	}

	// Get the file service for the current user
	h.fileService, err = h.authenticator.GetFileService()
	if err != nil {
		return err
	}

	return nil
}

// sendMessage sends a message to the client.
func (h *clientHandler) sendMessage(message interface{}) error {
	encoder := gob.NewEncoder(h.conn)
	return encoder.Encode(message)
}

// receiveMessage receives a message from the client.
func (h *clientHandler) receiveMessage(message interface{}) error {
	decoder := gob.NewDecoder(h.conn)
	return decoder.Decode(message)
}

// RegisterGobTypes registers the types that will be encoded and decoded using gob.
func RegisterGobTypes() {
	gob.Register(globalmodels.Message{})
	gob.Register(globalmodels.SyncRequest{})
	gob.Register(globalmodels.SyncResponse{})
	gob.Register(globalmodels.DownloadRequest{})
	gob.Register(globalmodels.UploadRequest{})
	gob.Register(globalmodels.FileChunk{})
	gob.Register(globalmodels.CloseConnection{})
	gob.Register(globalmodels.Error{})
}
