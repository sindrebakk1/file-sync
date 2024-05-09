package handlers

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleChunk is a mux.HandlerFunc
func HandleChunk(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleChunk")
	return nil
}
