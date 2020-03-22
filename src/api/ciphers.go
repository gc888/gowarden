package api

import (
	"encoding/json"
	"net/http"

	"github.com/404cn/gowarden/ds"
)

// Handle add ciphers.
func (apiHandler *APIHandler) HandleCiphers(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	apiHandler.logger.Infof("%v is trying add cipher.\n", email)

	// TODO acc
	_, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error("Failer to get account.")
		apiHandler.logger.Error(err)
		// TODO response writer
		return
	}

	var cipher ds.Cipher
	err = json.NewDecoder(r.Body).Decode(&cipher)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
	}
	defer r.Body.Close()

	var b []byte
	// TODO encode responss cipher to b
	b, err = json.Marshal(&cipher)
	if err != nil {
		apiHandler.logger.Error("Failed to encode json.")
		apiHandler.logger.Error(err)
		// TODO response writer
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
