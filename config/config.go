package config

import (
	"strings"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

var Config = ConfigParams{}

// Config parameters for torus metadata.
type ConfigParams struct {
	IPFSURL        string `env:"TM_IPFS_URL" envDefault:"localhost:5001"`
	Port           string `env:"TM_PORT" envDefault:"5051"`
	MySQLHostRead  string `env:"TM_MYSQL_HOST_READ"`
	MySQLHostWrite string `env:"TM_MYSQL_HOST_WRITE"`
	MySQLPort      string `env:"TM_MYSQL_PORT"`
	MySQLUser      string `env:"TM_MYSQL_USER"`
	MySQLDBName    string `env:"TM_MYSQL_DBNAME"`
	MySQLPassword  string `env:"TM_MYSQL_PASSWORD"`
	Debug          bool   `env:"TM_DEBUG"`
	HTTPSEnabled   bool   `env:"TM_HTTPS_ENABLED"`
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
