package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"strings"

	"encoding/json"

	"github.com/404cn/gowarden/sqlite/mock"
)

var mux *http.ServeMux
var writer *httptest.ResponseRecorder
var testHandler = New(mock.New())

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	mux = http.NewServeMux()
	mux.HandleFunc("/api/accounts/register", testHandler.HandleRegister)
	mux.HandleFunc("/api/accounts/prelogin", testHandler.HandlePrelogin)
	mux.HandleFunc("/identity/connect/token", testHandler.HandleLogin)
	writer = httptest.NewRecorder()
}

func tearDown() {

}

func TestHandleLogin(t *testing.T) {
	// TODO test login
	// TODO test refresh token
}

func TestHandlePrelogin(t *testing.T) {
	got := strings.NewReader(`{"email": "nobody@example.com"}`)

	request, _ := http.NewRequest("POST", "/api/accounts/prelogin", got)
	mux.ServeHTTP(writer, request)

	var want struct{ Kdf, KdfIterations int }
	json.Unmarshal(writer.Body.Bytes(), &want)

	if want.Kdf != 0 || want.KdfIterations != 100000 {
		t.Errorf("kdf: %v, kdfIterations: %v", want.Kdf, want.KdfIterations)
	}
}

func TestHandleRegister(t *testing.T) {
	got := strings.NewReader(`{
"name": "",
"email": "nobody@example.com",
"masterPasswordHash": "kuz4if+vSRXH+bCYLRyN6QonjvA5YglyUGW9/CI0Vqc=",
"masterPasswordHint": "example",
"key": "test_key",
"kdf": 0,
"kdfIterations": 100000,
"keys": {
"publicKey": "test_public_key",
"encryptedPrivateKey": "test_encrypted_private_key"
}
}`)

	request, _ := http.NewRequest("POST", "/api/accounts/register", got)
	mux.ServeHTTP(writer, request)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}
