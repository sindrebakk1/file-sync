package main

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"server/pkg/auth"
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
	BaseDir       = "files"
	ChallengeLen  = 32
)

func main() {
	log.Info("Starting server...")
	var (
		server             fileserver.FileServer
		userService        services.UserService
		fileServiceFactory services.FileServiceFactory
		fileCache          cache.Cache
		metaCache          cache.Cache
		cert               tls.Certificate
		tlsConfig          *tls.Config
		authConfig         *auth.Config
		err                error
	)

	// Initialize services.
	fileServiceFactory = services.NewFileServiceFactory(BaseDir)
	userService = services.NewUserService(fileServiceFactory)
	if err != nil {
		log.Fatal(err)
	}
	fileCache = cache.NewCache(FileCacheSize)
	metaCache = cache.NewCache(MetaCacheSize)

	// Load the TLS certificate.
	cert, err = tls.LoadX509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the file server.
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	authConfig = &auth.Config{
		ChallengeLen: ChallengeLen,
	}
	server, err = fileserver.NewFileServer(userService, fileCache, metaCache, tlsConfig, authConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Start the server.
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
