package main

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"server/pkg/auth"
	"server/pkg/cache"
	_ "server/pkg/cache"
	"server/pkg/mux"
	fileserver "server/pkg/server"
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
		server             fileserver.Server
		userService        services.UserService
		fileServiceFactory services.FileServiceFactory
		fileCache          cache.Cache
		metaCache          cache.Cache
		cert               tls.Certificate
		tlsConfig          *tls.Config
		authConfig         *auth.Config
		err                error
	)

	// Initialize userService and its dependencies.
	fileCache = cache.NewCache(FileCacheSize)
	metaCache = cache.NewCache(MetaCacheSize)
	fileServiceFactory = services.NewFileServiceFactory(BaseDir, fileCache, metaCache)
	userService = services.NewUserService(fileServiceFactory)
	if err != nil {
		log.Fatal(err)
	}

	authConfig = &auth.Config{
		ChallengeLen: ChallengeLen,
	}
	authenticator := auth.NewAuthenticator(userService, authConfig)

	clientMux := mux.NewMux(authenticator)

	// Load the TLS certificate.
	cert, err = tls.LoadX509KeyPair(CertFile, KeyFile)
	if err != nil {
		log.Fatal(err)
	}
	// Initialize the file server.
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server = fileserver.NewServer(clientMux, tlsConfig)

	// Start the server.
	err = server.ListenAndServe(Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Server stopped.")
}

func init() {
	// Configure logging.
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
}
