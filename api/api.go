package api

import (
	"crypto/sha256"
	"net/http"

	"encoding/base64"
	"strings"

	"encoding/json"

	"log"

	"errors"

	"github.com/404cn/gowarden/ds"
	"github.com/404cn/gowarden/sqlite"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
)

type handler interface {
	AddAccount(ds.Account) error
	GetAccount(string) (ds.Account, error)
}

type ApiHandler struct {
	db handler
}

func New(db handler) *ApiHandler {
	return &ApiHandler{
		db: db,
	}
}

var StdApiHandler = New(sqlite.StdDB)

func (apihandler *ApiHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	var err error
	r.ParseForm()

	email := r.PostForm["username"][0]
	password := r.PostForm["password"][0]

	log.Println(email + " is trying to login.")
	acc, err = checkPassword(email, password, apihandler.db)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		return
	}

	// TODO refresh token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.Claims{})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (apiHandler *ApiHandler) HandlePrelogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	acc, err = apiHandler.db.GetAccount(acc.Email)
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

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (apiHandler *ApiHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	if acc.KdfIterations < 5000 || acc.KdfIterations > 100000 {
		log.Println(err)
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

	err = apiHandler.db.AddAccount(acc)
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

func checkPassword(email, password string, db handler) (ds.Account, error) {
	acc, err := db.GetAccount(email)
	if err != nil {
		return ds.Account{}, err
	}

	passwordHash, _ := makeKey(password, acc.Email, acc.KdfIterations)
	if passwordHash != acc.MasterPasswordHash {
		return ds.Account{}, errors.New("Password wrong.")
	}

	return acc, nil
}
