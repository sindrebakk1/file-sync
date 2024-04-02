package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"server/routers"
)

func main() {
	fileRouter := routers.NewFileRouter()

	http.Handle("/file", fileRouter)

	// Start HTTP server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	log.Info("Server listening on :8080")
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
}
