package config

import (
	"strings"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

var Config = ConfigParams{}

// Config parameters for torus recoverer.
type ConfigParams struct {
	IPFSURL      string `env:"TR_IPFS_URL" envDefault:"localhost:5001"`
	Port         string `env:"TR_PORT" envDefault:"5051"`
	PGHost       string `env:"TR_PG_HOST"`
	PGPort       string `env:"TR_PG_PORT"`
	PGUser       string `env:"TR_PG_USER"`
	PGDBName     string `env:"TR_PG_DBNAME"`
	PGPassword   string `env:"TR_PG_PASSWORD"`
	Debug        bool   `env:"TR_DEBUG"`
	HTTPSEnabled bool   `env:"TR_HTTPS_ENABLED"`
}

func init() {
	conf := ConfigParams{}
	if err := env.Parse(&conf); err != nil {
		log.WithError(err).Fatal("could not parse config")
	}
	conf.Port = strings.Trim(conf.Port, "\"")
	Config = conf
}

func SetConfig(newConfig ConfigParams) {
	Config = newConfig
}
