package handlers

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleDelete is a mux.HandlerFunc
func HandleDelete(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleDelete")
	return nil
}
