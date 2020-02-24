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
	"github.com/404cn/gowarden/sqlite"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
)

const (
	jwtExpiresin  = 3600
	jwtSigningKey = "secret"
)

type handler interface {
	AddAccount(ds.Account) error
	GetAccount(string) (ds.Account, error)
	UpdateAccount(ds.Account) error
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

func (apiHandler *ApiHandler) HandleNegotiate(w http.ResponseWriter, r *http.Request) {
	email := GetEmailRctx(r)
}

func GetEmailRctx(r *http.Request) string {
	return r.Context().Value("email").(string)
}

func (apiHandler *ApiHandler) HandleSync(w http.ResponseWriter, r *http.Request) {

}

func (apiHandler *ApiHandler) HandleAccountKeys(w http.ResponseWriter, r *http.Request) {

}

func (apiHandler *ApiHandler) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth, ok := r.Header["Authorization"]
		if len(auth) < 1 && !ok {
			log.Println("No auth header.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			return
		}

		tokenString := strings.TrimPrefix(auth[0], "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Type assertion.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Signing method not right: %v", token.Header["alg"])
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSigningKey), nil
		})

		if nil != err {
			log.Println("JWT token parse error.")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
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
		w.Write([]byte(http.StatusText(401)))
	}
}

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

	if "" == acc.RefreshToken {
		acc.RefreshToken = createRefreshToken()
		err = apihandler.db.UpdateAccount(acc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Second * time.Duration(jwtExpiresin)).Unix(),
		"iss":     "gowarden",
		"sub":     "gowarden",
		"email":   acc.Email,
		"name":    acc.Name,
		"premium": true,
	})
	accessToken, err := token.SignedString([]byte(jwtSigningKey))
	if nil != err {
		log.Println(err)
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
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
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
