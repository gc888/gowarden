package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (apiHandler APIHandler) HandleFolderDelete(w http.ResponseWriter, r *http.Request) {
	folderUUID := mux.Vars(r)["folderUUID"]
	email := getEmailRctx(r)

	apiHandler.logger.Infof("%v is trying to delete a folder", email)

	err := apiHandler.db.DeleteFolder(folderUUID)
	if err != nil {
		apiHandler.logger.Error(err)
		apiHandler.logger.Error("Failed to delete folder.")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
}

func (apiHandler APIHandler) HandleFolderRename(w http.ResponseWriter, r *http.Request) {
	var rfolder struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&rfolder)
	if err != nil {
		apiHandler.logger.Error(err)
		apiHandler.logger.Error("Falied to decode json.")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	folderUUID := mux.Vars(r)["folderUUID"]
	email := getEmailRctx(r)

	apiHandler.logger.Infof("%v is trying to rename a folder", email)

	folder, err := apiHandler.db.RenameFolder(rfolder.Name, folderUUID)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	b, err := json.Marshal(&folder)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// handle add folers
func (apiHandler APIHandler) HandleFolder(w http.ResponseWriter, r *http.Request) {
	var rfolder struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&rfolder)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	emali := getEmailRctx(r)
	acc, err := apiHandler.db.GetAccount(emali)
	if err != nil {
		apiHandler.logger.Error("Can't get account.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	folder, err := apiHandler.db.AddFolder(acc.Id, rfolder.Name)
	if err != nil {
		apiHandler.logger.Error("Failed to add folder.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	b, err := json.Marshal(&folder)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
