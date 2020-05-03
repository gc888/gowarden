package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/404cn/gowarden/ds"
	jwt "github.com/dgrijalva/jwt-go"
)

// Handle login and refresh token.
// TODO refresh token timeout return 401.
func (apiHandler *APIHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var acc ds.Account
	var err error
	r.ParseForm()

	grantType, ok := r.PostForm["grant_type"]
	if !ok {
		apiHandler.logger.Error(err)
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
			apiHandler.logger.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}

		if refreshToken != acc.RefreshToken {
			apiHandler.logger.Error(err)
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
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	acc, err = apiHandler.db.GetAccount(acc.Email)
	if err != nil {
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
		apiHandler.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}
