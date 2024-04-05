package main

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"server/pkg/cache"
	"server/pkg/fileserver"
	"server/services"
)

const (
	Port          = 443
	FileCacheSize = 1_000
	MetaCacheSize = 100_000
	CertFile      = "certs/server.crt"
	KeyFile       = "certs/server.key"
	BaseDir       = "/data"
)

func main() {
	log.Info("Starting server...")
	var (
		service   services.FileService
		fileCache cache.Cache
		metaCache cache.Cache
		cert      tls.Certificate
		tlsConfig *tls.Config
		server    fileserver.FileServer
		err       error
	)
	service, err = services.NewFileService(BaseDir)
	if err != nil {
		log.Fatal(err)
	}
	fileCache = cache.NewCache(FileCacheSize)
	metaCache = cache.NewCache(MetaCacheSize)

	cert, err = tls.LoadX509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server, err = fileserver.NewFileServer(service, fileCache, metaCache, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	err = server.ListenAndServe(Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Server stopped.")
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
