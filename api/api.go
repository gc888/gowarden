package api

import (
	"crypto/sha256"
	"net/http"

	"encoding/base64"
	"strings"

	"encoding/json"

	"log"

	"github.com/404cn/gowarden/ds"
	"github.com/404cn/gowarden/sqlite"
	"golang.org/x/crypto/pbkdf2"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var login ds.Login

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&login)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}
}

func HandlePrelogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	acc, err = sqlite.StdDB.GetAccount(acc.Email)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	data := struct {
		// Must upper case so that can write into response.
		Kdf           int
		KdfIterations int
	}{
		Kdf:           acc.Kdf,
		KdfIterations: acc.KdfIterations,
	}

	d, err := json.Marshal(&data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	log.Println(d)

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	if acc.KdfIterations < 5000 || acc.KdfIterations > 100000 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	log.Println(acc.Email + " is trying to register.")

	acc.MasterPasswordHash, err = makeKey(acc.MasterPasswordHash, acc.Email, acc.KdfIterations)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	err = sqlite.StdDB.AddAccount(acc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func makeKey(password, salt string, iterations int) (string, error) {
	salt = strings.ToLower(salt)
	p, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", err
	}

	masterKey := pbkdf2.Key(p, []byte(salt), iterations, 256/8, sha256.New)

	return base64.StdEncoding.EncodeToString(masterKey), nil
}
