package handlers

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleUpload is a mux.HandlerFunc
func HandleUpload(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleUpload")
	return nil
}
