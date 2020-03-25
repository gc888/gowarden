package api

import (
	"encoding/json"
	"net/http"

	"github.com/404cn/gowarden/ds"
)

func (apiHandler *APIHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	if acc.KdfIterations < 5000 || acc.KdfIterations > 100000 {
		apiHandler.logger.Error("Bad kdf iterations.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	apiHandler.logger.Info(acc.Email + " is trying to register.")

	acc.MasterPasswordHash, err = makeKey(acc.MasterPasswordHash, acc.Email, acc.KdfIterations)
	if err != nil {
		apiHandler.logger.Error("Failed to generate key.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	err = apiHandler.db.AddAccount(acc)
	if err != nil {
		apiHandler.logger.Error("Failed to get account.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	w.WriteHeader(http.StatusOK)
}
