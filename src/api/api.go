package api

import (
	"net/http"

	"encoding/json"

	"github.com/404cn/gowarden/ds"
)

func (apiHandler *APIHandler) HandleNegotiate(w http.ResponseWriter, r *http.Request) {

}

func (apiHandler *APIHandler) HandleSync(w http.ResponseWriter, r *http.Request) {

}

// Update account's keys.
func (apiHandler *APIHandler) HandleAccountKeys(w http.ResponseWriter, r *http.Request) {
	var keys ds.Keys
	err := json.NewDecoder(r.Body).Decode(&keys)
	defer r.Body.Close()
	if nil != err {
		apiHandler.logger.Error("Failed to decode request body.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	email := getEmailRctx(r)

	acc, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Errorf("Failed to get account with %v", email)
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	acc.Keys = keys
	apiHandler.db.UpdateAccount(acc)
}
