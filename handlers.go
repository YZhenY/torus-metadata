package main

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/patrickmn/go-cache"
	"github.com/torusresearch/torus-metadata/config"
)

type (
	Data struct {
		gorm.Model
		Key   string `gorm:"unique;not null"`
		Value string
	}

	// SetHandler is a reference the database.
	SetHandler struct {
		sh      *shell.Shell
		db      *gorm.DB
		timeout time.Duration
		cache   *cache.Cache
		Debug   bool
	}
	// SetParams are the params needed for authorization.
	SetParams struct {
		Namespace string  `json:"namespace"`
		PubKeyX   big.Int `json:"pub_key_X"`
		PubKeyY   big.Int `json:"pub_key_Y"`
		SetData   SetData `json:"set_data"`
		Signature []byte  `json:"signature"`
	}
	SetData struct {
		Data      string  `json:"data"`
		Timestamp big.Int `json:"timestamp"`
	}
	// SetResult is the response of an authorization.
	SetResult struct {
		Message string `json:"message"`
	}

	// GetHandler is a reference the database.
	GetHandler struct {
		db *gorm.DB
	}
	// SetParams are the params needed for authorization.
	GetParams struct {
		Namespace string  `json:"namespace"`
		PubKeyX   big.Int `json:"pub_key_X"`
		PubKeyY   big.Int `json:"pub_key_Y"`
	}
	// SetResult is the response of an authorization.
	GetResult struct {
		Message string `json:"message"`
	}

	// Health endpoint
	HealthHandler struct {
	}
)

// SetupHTTPHandler registers the set and get handlers.
func SetupHTTPHandler(cfg config.ConfigParams) (*http.ServeMux, error) {
	mr := http.NewServeMux()
	sh := shell.NewShell(cfg.IPFSURL)
	dbRead, err := gorm.Open("mysql",
		fmt.Sprintf(
			"%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQLUser,
			cfg.MySQLPassword,
			cfg.MySQLHostRead,
			cfg.MySQLPort,
			cfg.MySQLDBName,
		),
	)
	dbWrite, err := gorm.Open("mysql",
		fmt.Sprintf(
			"%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQLUser,
			cfg.MySQLPassword,
			cfg.MySQLHostRead,
			cfg.MySQLPort,
			cfg.MySQLDBName,
		),
	)
	c := cache.New(10*time.Minute, 10*time.Minute)

	if err != nil {
		return nil, err
	}

	dbRead.LogMode(true)
	dbWrite.LogMode(true)
	dbWrite.AutoMigrate(&Data{})

	mr.Handle("/set", SetHandler{sh: sh, db: dbWrite, Debug: cfg.Debug, timeout: time.Minute, cache: c})
	mr.Handle("/get", GetHandler{db: dbRead})
	mr.Handle("/health", HealthHandler{})
	return mr, nil
}
