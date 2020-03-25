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

	acc, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error("Failed to get account.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	var cipher ds.Cipher
	err = json.NewDecoder(r.Body).Decode(&cipher)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	resCipher, err := apiHandler.db.AddCipher(cipher, acc.Id)
	if err != nil {
		// TODO
	}

	var b []byte
	// TODO encode responss cipher to b
	b, err = json.Marshal(&resCipher)
	if err != nil {
		apiHandler.logger.Error("Failed to encode json.")
		apiHandler.logger.Error(err)
		// TODO response writer
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (apiHander *APIHandler) HandleUpdateCiphers(w http.ResponseWriter, r *http.Request) {

}

func (apiHander *APIHandler) HandleDeleteCiphers(w http.ResponseWriter, r *http.Request) {

}
