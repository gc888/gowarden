package api

import (
	"crypto/sha256"
	"net/http"

	"encoding/base64"
	"strings"

	"encoding/json"

	"log"

	"github.com/404cn/gowarden/database"
	"github.com/404cn/gowarden/ds"
	"golang.org/x/crypto/pbkdf2"
)

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

	log.Println(acc.Email + "is trying to register.")

	// Just for test.
	// TODO delete
	acc.ToString()

	acc.MasterPasswordHash, err = makeKey(acc.MasterPasswordHash, acc.Email, acc.KdfIterations)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	err = database.StdDB.AddAccount(acc)
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
