package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/404cn/gowarden/ds"
)

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

	// TODO
	var cipherForUpdate ds.CipherForUpdate

	err = json.NewDecoder(r.Body).Decode(&cipherForUpdate)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer r.Body.Close()

	// TODO
	cipher.Type, cipher.FolderId, cipher.OrganizationId = cipherForUpdate.Type, cipherForUpdate.FolderId, cipherForUpdate.OrganizationId
	cipher.Name, cipher.Notes, cipher.Favorite = cipherForUpdate.Name, cipherForUpdate.Notes, cipherForUpdate.Favorite
	cipher.Login, cipher.Fields = cipherForUpdate.Login, cipherForUpdate.Fields
	cipher.Card, cipher.Identity, cipher.SecureNote = cipherForUpdate.Card, cipherForUpdate.Identity, cipherForUpdate.SecureNote

	cipher.Id = cipherId
	cipher, err = apiHandler.db.UpdateCipher(cipher, acc.Id)
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

	err = os.RemoveAll("attachments/" + cipherId)
	if err != nil {
		apiHandler.logger.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	apiHandler.logger.Infof("Cipher %v deleted.", cipherId)
	return
}

func (apiHandler APIHandler) HandleAddAttachment(w http.ResponseWriter, r *http.Request) {
	var attachment ds.Attachment
	email := getEmailRctx(r)
	cipherId := mux.Vars(r)["cipherId"]

	attachment.Id = uuid.Must(uuid.NewRandom()).String()

	// TODO replace server and port to gowaden's flag
	// FIXME usl must be http://server:port/attachments/{cipherid}/{attachmentid}
	attachment.Url = strings.ToLower(strings.Split(r.Proto, "/")[0]) + "://" + r.Host + "/attachments/" + cipherId + "/" + attachment.Id

	apiHandler.logger.Infof("%v is trying to add attachment.", email)

	parseErr := r.ParseMultipartForm(0)
	if parseErr != nil {
		apiHandler.logger.Error(parseErr)
		http.Error(w, "failed to parse multipart message", http.StatusBadRequest)
		return
	}

	attachment.Key = r.FormValue("key")

	for _, h := range r.MultipartForm.File["data"] {
		attachment.FileName = h.Filename
		attachment.Size = strconv.FormatInt(h.Size, 10)

		file, _ := h.Open()

		_, err := os.Stat("attachments/" + cipherId)
		if err != nil {
			apiHandler.logger.Info("Didn't find cipher's folder, try to create.")
			os.Mkdir("attachments/"+cipherId, os.ModePerm)
		}

		tmpfile, err := os.Create(regexp.MustCompile(`attachments.*`).FindStringSubmatch(attachment.Url)[0])
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

	apiHandler.logger.Infof("%v is trying to delete attachment.\n", email)

	url, err := apiHandler.db.DeleteAttachment(cipherId, attachmentId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	err = os.Remove(regexp.MustCompile(`attachments.*`).FindStringSubmatch(url)[0])
	if err != nil {
		apiHandler.logger.Error(err)
	}

	return
}

func (apiHandler APIHandler) HandleGetAttachment(w http.ResponseWriter, r *http.Request) {
	cipherId := mux.Vars(r)["cipherId"]
	attachmentId := mux.Vars(r)["attachmentId"]

	apiHandler.logger.Infof("trying to download attachment: %v.", attachmentId)

	attachment, err := apiHandler.db.GetAttachment(cipherId, attachmentId)
	if err != nil {
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	// TODO doesn't work
	http.ServeFile(w, r, regexp.MustCompile(`attachment.*`).FindStringSubmatch(attachment.Url)[0])
}
