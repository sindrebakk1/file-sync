package main

import (
	"crypto/tls"
	"errors"
	"file-sync/constants"
	"file-sync/enums"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path/filepath"
	"server/pkg/cache"
	"server/pkg/fileserver"
	"server/pkg/mux"
	"server/services"
)

var (
	Environment   enums.Environment
	Port          int
	FileCacheSize int
	MetaCacheSize int
	CertDir       string
	CertFile      string
	KeyFile       string
	BaseDir       string
	ChallengeLen  int
	LogLevel      log.Level
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
		authConfig         *services.Config
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

	authConfig = &services.Config{
		ChallengeLen: ChallengeLen,
	}
	authenticator := services.NewAuthService(userService, authConfig)

	tcpMux := mux.NewMux(authenticator)

	// Load the TLS certificate.
	var certDir string
	certDir, err = filepath.Abs(CertDir)
	if err != nil {
		log.Fatal(err)
	}
	cert, err = tls.LoadX509KeyPair(filepath.Join(certDir, CertFile), filepath.Join(certDir, KeyFile))
	if err != nil {
		log.Fatal(err)
	}
	// Initialize the file server.
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server = fileserver.NewServer(tcpMux, tlsConfig)

	// Start the server.
	err = server.ListenAndServe(Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Server stopped.")
}

func init() {
	// Set default values.
	viper.SetDefault("env", enums.Development)
	viper.SetDefault("port", 443)
	viper.SetDefault("cache.file.size", 1_000)
	viper.SetDefault("cache.meta.size", 100_000)
	viper.SetDefault("tls.dir", "./certs")
	viper.SetDefault("tls.cert", "server.crt")
	viper.SetDefault("tls.key", "server.key")
	viper.SetDefault("data.dir", "./data")
	viper.SetDefault("auth.challenge.len", 32)
	viper.SetDefault("log.level", log.ErrorLevel)

	// Configure environment variables.
	viper.SetEnvPrefix(constants.AppName)
	viper.AutomaticEnv()

	Environment = enums.Environment(viper.GetString("env"))

	// Set the configuration file name based on the environment.
	var configName string
	switch Environment {
	case enums.Production:
		configName = "config"
	case enums.Development:
		configName = "config.dev"
	default:
		log.Fatalf("invalid environment: %s", Environment)
	}

	// Set the configuration file name and path.
	viper.SetConfigName(configName)
	viper.SetConfigType("json")
	viper.AddConfigPath(fmt.Sprintf("/etc/%s", constants.AppName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", constants.AppName))
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Fatal(err)
		}
	}

	// Bind environment variables to variables.
	Port = viper.GetInt("port")
	FileCacheSize = viper.GetInt("cache.file.size")
	MetaCacheSize = viper.GetInt("cache.meta.size")
	CertDir = viper.GetString("tls.dir")
	CertFile = viper.GetString("tls.cert")
	KeyFile = viper.GetString("tls.key")
	BaseDir = viper.GetString("data.dir")
	ChallengeLen = viper.GetInt("auth.challenge.len")
	LogLevel = viper.Get("log.level").(log.Level)

	// Configure logging.
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(LogLevel)
}
