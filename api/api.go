package api

import (
	"crypto/sha256"
	"net/http"

	"encoding/base64"
	"strings"

	"encoding/json"

	"log"

	"errors"

	"time"

	"crypto/rand"

	"context"

	"fmt"

	"github.com/404cn/gowarden/ds"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/crypto/pbkdf2"
)

const (
	jwtExpiresin = 3600
)

type handler interface {
	AddAccount(ds.Account) error
	GetAccount(string) (ds.Account, error)
	UpdateAccount(ds.Account) error
	AddFolder(string, string) (ds.Folder, error)
	DeleteFolder(string) error
	RenameFolder(string, string) (ds.Folder, error)
}

type APIHandler struct {
	db         handler
	signingKey string
	logger     *zap.SugaredLogger
}

func New(db handler, key string, sugar *zap.SugaredLogger) *APIHandler {
	return &APIHandler{
		db:         db,
		signingKey: key,
		logger:     sugar,
	}
}

func (apiHandler *APIHandler) HandleNegotiate(w http.ResponseWriter, r *http.Request) {

}

func (apiHandler *APIHandler) HandleSync(w http.ResponseWriter, r *http.Request) {

}

func (apiHandler APIHandler) HandleFolderDelete(w http.ResponseWriter, r *http.Request) {
	folderUUID := mux.Vars(r)["folderUUID"]
	email := GetEmailRctx(r)

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
	email := GetEmailRctx(r)

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

	emali := GetEmailRctx(r)
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

// Handle add ciphers.
func (apiHandler *APIHandler) HandleCiphers(w http.ResponseWriter, r *http.Request) {
	email := GetEmailRctx(r)
	apiHandler.logger.Infof("%v is trying add cipher.\n", email)

	// TODO acc
	_, err := apiHandler.db.GetAccount(email)
	if nil != err {
		apiHandler.logger.Error("Failer to get account.")
		apiHandler.logger.Error(err)
		// TODO response writer
		return
	}

	var cipher ds.Cipher
	err = json.NewDecoder(r.Body).Decode(&cipher)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
	}
	defer r.Body.Close()

	var b []byte
	// TODO encode responss cipher to b
	b, err = json.Marshal(&cipher)
	if err != nil {
		apiHandler.logger.Error("Failed to encode json.")
		apiHandler.logger.Error(err)
		// TODO response writer
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
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

	email := GetEmailRctx(r)

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

// Return request context's email value.
func GetEmailRctx(r *http.Request) string {
	return r.Context().Value("email").(string)
}

// Middleware to handle login auth.
func (apiHandler *APIHandler) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth, ok := r.Header["Authorization"]
		if len(auth) < 1 && !ok {
			apiHandler.logger.Error("No auth header.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		tokenString := strings.TrimPrefix(auth[0], "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Type assertion.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				apiHandler.logger.Errorf("Signing method not right: %v\n", token.Header["alg"])
				return nil, fmt.Errorf("Unexpected signing method: %v\n", token.Header["alg"])
			}
			return []byte(apiHandler.signingKey), nil
		})

		if nil != err {
			apiHandler.logger.Error("JWT token parse error.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email, ok := claims["email"].(string)
			if ok {
				// Add email to request's context so that can get account by email.
				ctx := context.WithValue(r.Context(), "email", email)
				h(w, r.WithContext(ctx))
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
	}
}

// Handle login and refresh token.
// TODO refresh token timeout return 401.
func (apiHandler *APIHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	var err error
	r.ParseForm()

	grantType, ok := r.PostForm["grant_type"]
	if !ok {
		apiHandler.logger.Error("No grant type.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	if grantType[0] == "refresh_token" {
		refreshToken := r.PostForm["refresh_token"][0]
		if len(refreshToken) != 32 {
			// TODO length 44, base64 encoded
			apiHandler.logger.Errorf("Bad token length: %v", len(refreshToken))
		}

		acc, err := apiHandler.db.GetAccount(refreshToken)
		if nil != err {
			apiHandler.logger.Error("Failed to get account from refresh token.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		if refreshToken != acc.RefreshToken {
			apiHandler.logger.Error("Bad refresh token.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		apiHandler.logger.Infof("%v is trying to refresh a token.\n", acc.Email)

	} else {
		// login in with email.
		email := r.PostForm["username"][0]
		password := r.PostForm["password"][0]

		apiHandler.logger.Info(email + " is trying to login.")
		acc, err = checkPassword(email, password, apiHandler.db)
		if err != nil {
			apiHandler.logger.Error("Incorrect password.")
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

	}

	// If accounts refresh token is not empty, do not change it or the other clients will be logged out.
	if "" == acc.RefreshToken {
		acc.RefreshToken = createRefreshToken()
		err = apiHandler.db.UpdateAccount(acc)
		if err != nil {
			// FIXME jwt token timeout 401
			apiHandler.logger.Error("Failed to update account info.")
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
	}

	// Gen a  jwt as access token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Second * time.Duration(jwtExpiresin)).Unix(),
		"iss":     "gowarden",
		"sub":     "gowarden",
		"email":   acc.Email,
		"name":    acc.Name,
		"premium": true,
	})
	accessToken, err := token.SignedString([]byte(apiHandler.signingKey))
	if nil != err {
		apiHandler.logger.Error("Failed to signing jwt token.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	rtoken := struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		Key          string `json:"Key"`
	}{
		AccessToken:  accessToken,
		ExpiresIn:    jwtExpiresin,
		TokenType:    "Bearer",
		RefreshToken: acc.RefreshToken,
		Key:          acc.Key,
	}

	d, err := json.Marshal(&rtoken)
	if nil != err {
		apiHandler.logger.Error("Failed to marshal json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (apiHandler *APIHandler) HandlePrelogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	acc, err = apiHandler.db.GetAccount(acc.Email)
	if err != nil {
		apiHandler.logger.Error("Failed to get account.")
		apiHandler.logger.Error(err)
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
		apiHandler.logger.Error("Failed to marshal json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}

func (apiHandler *APIHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&acc)
	if err != nil {
		apiHandler.logger.Error("Failed to decode json.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	if acc.KdfIterations < 5000 || acc.KdfIterations > 100000 {
		apiHandler.logger.Error("Bad kdf iterations.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	apiHandler.logger.Info(acc.Email + " is trying to register.")

	acc.MasterPasswordHash, err = makeKey(acc.MasterPasswordHash, acc.Email, acc.KdfIterations)
	if err != nil {
		apiHandler.logger.Error("Failed to generate key.")
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	err = apiHandler.db.AddAccount(acc)
	if err != nil {
		apiHandler.logger.Error("Failed to get account.")
		apiHandler.logger.Error(err)
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
