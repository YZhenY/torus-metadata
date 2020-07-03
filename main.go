package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/torus-metadata/config"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	log.WithFields(log.Fields{
		"Version": "1.1",
		"Port":    config.Config.Port,
		"Debug":   config.Config.Debug,
		"HTTPS":   config.Config.HTTPSEnabled,
	}).Info("config")

	mr, err := SetupHTTPHandler(config.Config)
	if err != nil {
		log.WithError(err).Fatal("could not start app")
	}

	handler := cors.New(cors.Options{MaxAge: 600, AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "x-api-key", "Authorization"}}).Handler(mr)
	listenString := fmt.Sprintf(":%s", config.Config.Port)

	if !config.Config.HTTPSEnabled {
		if err := http.ListenAndServe(listenString, handler); err != nil {
			log.WithError(err).Fatal()
		}
	} else {
		path, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.WithError(err).Fatal()
		}
		certFile := fmt.Sprintf("%s/ssl/server.crt", path)
		keyFile := fmt.Sprintf("%s/ssl/server.key", path)
		if err := http.ListenAndServeTLS(listenString, certFile, keyFile, handler); err != nil {
			log.WithError(err).Fatal()
		}
	}
}
