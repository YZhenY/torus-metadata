package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/torusresearch/bijson"
	"github.com/torusresearch/torus-common/common"
	torusCrypto "github.com/torusresearch/torus-common/crypto"
)

const contentType = "application/json; charset=utf-8"

type VerifierLookupParams struct {
	Verifier   string `json:"verifier"`
	VerifierID string `json:"verifier_id"`
}
type VerifierLookupItem struct {
	KeyIndex string  `json:"key_index"`
	PubKeyX  big.Int `json:"pub_key_X"`
	PubKeyY  big.Int `json:"pub_key_Y"`
	Address  string  `json:"address"`
}
type VerifierLookupResult struct {
	Keys []VerifierLookupItem `json:"keys"`
}

type verifierLookupRequestBody struct {
	RPCVersion string `json:"jsonrpc"`
	Method     string `json:"method"`
	ID         int    `json:"id"`

	Params VerifierLookupParams `json:"params"`
}
type verifierLookupResponse struct {
	ID         int                  `json:"id"`
	RPCVersion string               `json:"jsonrpc"`
	Result     VerifierLookupResult `json:"result"`
}

type ErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type ErrorWrapper struct {
	Error ErrorObject `json:"error"`
}

func getRespError(resp []byte) error {
	var errWrapper ErrorWrapper
	if err := bijson.Unmarshal(resp, &errWrapper); err != nil {
		return nil
	}

	if errObject := errWrapper.Error; errObject.Code < 0 && errObject.Message != "" {
		return fmt.Errorf("[%d] %s %s", errObject.Code, errObject.Message, errObject.Data)
	}

	return nil
}

func postRPC(endpoint string, client *http.Client, marshaledBody []byte, v interface{}) error {
	resp, err := client.Post(endpoint, contentType, bytes.NewBuffer(marshaledBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := getRespError(respBody); err != nil {
		return err
	}

	return bijson.Unmarshal(respBody, &v)
}

// ServeHTTP serves the HTTP for SetHandler.
func (h SetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var p SetParams
	if err = bijson.Unmarshal(params, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !h.cfg.Debug {

		pubKey := common.Point{
			X: p.PubKeyX,
			Y: p.PubKeyY,
		}

		bytesToVerify, err := bijson.Marshal(p.SetData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hexBytesToVerify := hex.EncodeToString(bytesToVerify)

		if _, found := h.cache.Get(hexBytesToVerify); found {
			http.Error(w, "message already seen in cache", http.StatusInternalServerError)
			return
		}

		h.cache.SetDefault(hexBytesToVerify, true)

		timeSigned := time.Unix(p.SetData.Timestamp.Int64(), 0)
		if timeSigned.Add(h.timeout).Before(time.Now()) {
			http.Error(w, "timesigned is more than 60 seconds ago", http.StatusInternalServerError)
			return
		}

		if !torusCrypto.VerifyPtFromRaw(bytesToVerify, pubKey, p.Signature) {
			http.Error(w, "invalid signature", http.StatusInternalServerError)
			return
		}
	}

	key := p.PubKeyX.Text(16) + "\x1c" + p.PubKeyY.Text(16)
	if p.Namespace != "" {
		key = key + "\x1c" + p.Namespace
	}

	data := Data{
		Key:   key,
		Value: p.SetData.Data,
	}
	// get hash only
	cid, err := h.sh.Add(strings.NewReader(p.SetData.Data), shell.OnlyHash(true))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// upload to ipfs in parallel
	go func(data string) {
		_, err := h.sh.Add(strings.NewReader(data))
		if err != nil {
			log.WithField("data", data).Error("could not ipfs add")
		}
	}(p.SetData.Data)

	if !strings.Contains(h.cfg.MySQLHostWrite, "ap-southeast-1") {
		h.db.Exec("SET aurora_replica_read_consistency='SESSION';")
	}

	if err := h.db.Where(Data{Key: data.Key}).Assign(Data{Value: data.Value}).FirstOrCreate(&data).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	result, err := bijson.Marshal(SetResult{Message: cid})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(result)
}

// ServeHTTP serves the HTTP for GetHandler.
func (h GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var p GetParams
	if err = bijson.Unmarshal(params, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := p.PubKeyX.Text(16) + "\x1c" + p.PubKeyY.Text(16)
	if p.Namespace != "" {
		key = key + "\x1c" + p.Namespace
	}
	var value Data
	h.db.Where(&Data{Key: key}).First(&value)
	result, err := bijson.Marshal(GetResult{Message: value.Value})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(result)
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "Ok!")
}
