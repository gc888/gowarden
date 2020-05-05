package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/404cn/gowarden/ds"
)

// Handle add ciphers.
func (apiHandler *APIHandler) HandleCiphers(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	apiHandler.logger.Infof("%v is trying add cipher.\n", email)

	acc, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	var cipher ds.Cipher
	err = json.NewDecoder(r.Body).Decode(&cipher)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	resCipher, err := apiHandler.db.AddCipher(cipher, acc.Id)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	var b []byte
	b, err = json.Marshal(&resCipher)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (apiHandler *APIHandler) HandleUpdateCiphers(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	apiHandler.logger.Infof("%v is trying to update cipher.", email)

	acc, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	cipherId := mux.Vars(r)["cipherId"]

	var cipher ds.Cipher
	err = json.NewDecoder(r.Body).Decode(&cipher)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	cipher.Id = cipherId
	err = apiHandler.db.UpdateCipher(cipher, acc.Id)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	d, err := json.Marshal(&cipher)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)

	apiHandler.logger.Infof("cipher %v updated.", cipherId)
	return
}

func (apiHandler *APIHandler) HandleDeleteCiphers(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	apiHandler.logger.Infof("%v is trying to delete cipher.", email)

	acc, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	cipherId := mux.Vars(r)["cipherId"]

	err = apiHandler.db.DeleteCipher(acc.Id, cipherId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	apiHandler.logger.Infof("Cipher %v deleted.", cipherId)
	return
}
