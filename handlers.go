package main

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
		Debug   bool
	}
	// SetParams are the params needed for authorization.
	SetParams struct {
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
		PubKeyX big.Int `json:"pub_key_X"`
		PubKeyY big.Int `json:"pub_key_Y"`
	}
	// SetResult is the response of an authorization.
	GetResult struct {
		Message string `json:"message"`
	}
)

// SetupHTTPHandler registers the set and get handlers.
func SetupHTTPHandler(cfg config.ConfigParams) (*http.ServeMux, error) {
	mr := http.NewServeMux()
	sh := shell.NewShell(cfg.IPFSURL)
	db, err := gorm.Open("postgres",
		fmt.Sprintf(
			"host='%s' port=%s user=%s dbname=%s password=%s",
			cfg.PGHost,
			cfg.PGPort,
			cfg.PGUser,
			cfg.PGDBName,
			cfg.PGPassword,
		),
	)

	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	db.AutoMigrate(&Data{})

	mr.Handle("/set", SetHandler{sh: sh, db: db, Debug: cfg.Debug, timeout: time.Minute})
	mr.Handle("/get", GetHandler{db: db})
	return mr, nil
}
