package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"strings"

	"encoding/json"

	"github.com/404cn/gowarden/logger"
	"github.com/404cn/gowarden/sqlite/mock"
)

var logT, _ = logger.New(5)
var muxForTest *http.ServeMux
var writer *httptest.ResponseRecorder
var testHandler = New(mock.New(), "", logT, "")

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	muxForTest = http.NewServeMux()
	muxForTest.HandleFunc("/api/accounts/register", testHandler.HandleRegister)
	muxForTest.HandleFunc("/api/accounts/prelogin", testHandler.HandlePrelogin)
	muxForTest.HandleFunc("/identity/connect/token", testHandler.HandleLogin)
	writer = httptest.NewRecorder()
}

func tearDown() {

}

func TestHandleLogin(t *testing.T) {
	reader := strings.NewReader(`
"grant_type" = "password",
"username" = "nobody@example.com",
"password" = "kuz4if+vSRXH+bCYLRyN6QonjvA5YglyUGW9/CI0Vqc=",
"scope" = "api offline_access",
"client_id" = "desktop",
"deviceType" = "7",
"deviceIdentifier" = "b299991b-f039-45b6-a350-2bbcdadaa37d",
"deviceName" = "macos",
`)

	req, _ := http.NewRequest("POST", "/identity/connect/token", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	muxForTest.ServeHTTP(writer, req)

	//if writer.Code != 200 {
	//t.Errorf("Response code is %v", writer.Code)
	//}
}

func TestHandlePrelogin(t *testing.T) {
	got := strings.NewReader(`{"email": "nobody@example.com"}`)

	request, _ := http.NewRequest("POST", "/api/accounts/prelogin", got)
	muxForTest.ServeHTTP(writer, request)

	var want struct{ Kdf, KdfIterations int }
	json.Unmarshal(writer.Body.Bytes(), &want)

	//if want.Kdf != 0 || want.KdfIterations != 100000 {
	//t.Errorf("kdf: %v, kdfIterations: %v", want.Kdf, want.KdfIterations)
	//}
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
	muxForTest.ServeHTTP(writer, request)

	//if writer.Code != 200 {
	//t.Errorf("Response code is %v", writer.Code)
	//}
}

func BenchmarkRegister(b *testing.B) {
	var t *testing.T
	for i := 0; i < b.N; i++ {
		TestHandleRegister(t)
	}
}

func BenchmarkPrelogin(b *testing.B) {
	var t *testing.T
	for i := 0; i < b.N; i++ {
		TestHandlePrelogin(t)
	}
}

func BenchmarkLogin(b *testing.B) {
	var t *testing.T
	for i := 0; i < b.N; i++ {
		TestHandleLogin(t)
	}
}
