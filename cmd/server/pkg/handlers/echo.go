package handlers

import (
	"filesync/enums"
	"filesync/models"
	"server/pkg/mux"
)

// HandleEcho is a mux.HandlerFunc
func HandleEcho(resChan chan models.Message, req *mux.Request) error {
	message := req.Message
	message.Header.Sender = enums.Server
	resChan <- message
	return nil
}
