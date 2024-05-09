package handlers

import (
	"filesync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleStatus is a mux.HandlerFunc
func HandleStatus(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleStatus")
	return nil
}
