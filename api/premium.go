package api

import (
	"encoding/json"
	"github.com/404cn/gowarden/ds"
	"github.com/gorilla/mux"
	"net/http"
)

func (apiHandler APIHandler) HandleAddAttachment(w http.ResponseWriter, r *http.Request) {
	var attachment ds.Attachment
	email := getEmailRctx(r)
	cipherId := mux.Vars(r)["cipherId"]

	apiHandler.logger.Info("%v is trying to add attachment.\n", email)

	parseErr := r.ParseMultipartForm(0)
	if parseErr != nil {
		apiHandler.logger.Error(parseErr)
		http.Error(w, "failed to parse multipart message", http.StatusBadRequest)
		return
	}

	attachment.Key = r.FormValue("key")

	for _, h := range r.MultipartForm.File["data"] {
		attachment.FileName = h.Filename
		attachment.Size = h.Size
	}

	cipher, err := apiHandler.db.AddAttachment(cipherId, attachment)
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
	return
}

func (apiHandler APIHandler) HandleDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	cipherId := mux.Vars(r)["cipherId"]
	attachmentId := mux.Vars(r)["attachmentId"]

	apiHandler.logger.Info("%v is trying to delete attachment: %v.\n", email, attachmentId)

	err := apiHandler.db.DeleteAttachment(cipherId, attachmentId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	return
}

// TODO download attachments
func (apiHandler APIHandler) HandleGetAttachment(w http.ResponseWriter, r *http.Request) {

}
