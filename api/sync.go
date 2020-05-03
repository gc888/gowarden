package api

import (
	"encoding/json"
	"github.com/404cn/gowarden/ds"
	"net/http"
)

func (apiHandler APIHandler) HandleSync(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)

	apiHandler.logger.Infof("%v is trying to sync.", email)

	acc, err := apiHandler.db.GetAccount(email)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	// TODO: Profile, Folders, Ciphers, Domains
	profile := acc.Profile()

	// TODO cipher.data 在 database 中都是null
	ciphers, err := apiHandler.db.GetCiphers(acc.Id)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	folders, err := apiHandler.db.GetFolders(acc.Id)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	domains := ds.Domains{
		EquivalentDomains:       nil,
		GlobalEquivalentDomains: nil,
		Object:                  "domains",
	}

	data := ds.SyncData{
		Profile: profile,
		Folders: folders,
		Ciphers: ciphers,
		Domains: domains,
		Object:  "sync",
	}

	jsonData, err := json.Marshal(&data)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
