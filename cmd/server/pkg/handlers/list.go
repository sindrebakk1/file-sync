package handlers

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleList is a mux.HandlerFunc
func HandleList(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleList")
	return nil
}
