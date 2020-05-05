package api

import (
	"encoding/json"
	"github.com/404cn/gowarden/ds"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func (apiHandler APIHandler) HandleAddAttachment(w http.ResponseWriter, r *http.Request) {
	var attachment ds.Attachment
	email := getEmailRctx(r)
	cipherId := mux.Vars(r)["cipherId"]

	attachment.Id = uuid.Must(uuid.NewRandom()).String()
	attachment.Url = "attachments/" + cipherId + "/" + attachment.Id

	apiHandler.logger.Infof("%v is trying to add attachment.\n", email)

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

		file, _ := h.Open()

		_, err := os.Stat("attachments/" + cipherId)
		if err != nil {
			apiHandler.logger.Info("Didn't find cipher's folder, try to create.")
			os.Mkdir("attachments/"+cipherId, os.ModePerm)
		}

		tmpfile, err := os.Create(attachment.Url)
		if err != nil {
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
		defer tmpfile.Close()
		io.Copy(tmpfile, file)
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

	apiHandler.logger.Infof("%v is trying to delete attachment: %v.\n", email, attachmentId)

	url, err := apiHandler.db.DeleteAttachment(cipherId, attachmentId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	err = os.Remove(url)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	return
}

// TODO download attachments
// FIXME didn't get client's request
func (apiHandler APIHandler) HandleGetAttachment(w http.ResponseWriter, r *http.Request) {
	email := getEmailRctx(r)
	cipherId := mux.Vars(r)["cipherId"]
	attachmentId := mux.Vars(r)["attachmentId"]

	apiHandler.logger.Infof("%v is trying to download attachment: %v.\n", email, attachmentId)

	attachment, err := apiHandler.db.GetAttachment(cipherId, attachmentId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	// TODO
	file, _ := os.Open(attachment.Url)
	defer file.Close()
	b, _ := ioutil.ReadAll(file)

	w.Write(b)

	return
}
