package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"strings"

	"net/http"

	"github.com/404cn/gowarden/ds"
	"golang.org/x/crypto/pbkdf2"
)

// getEmailRctx return email from request's context
func getEmailRctx(r *http.Request) string {
	return r.Context().Value("email").(string)
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
	if nil != err {
		return ds.Account{}, err
	}

	passwordHash, _ := makeKey(password, acc.Email, acc.KdfIterations)
	if passwordHash != acc.MasterPasswordHash {
		return ds.Account{}, errors.New("Password wrong.")
	}

	return acc, nil
}

// Generate 32 bit rand string.
func createRefreshToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if nil != err {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
