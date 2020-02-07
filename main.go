package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/pbkdf2"
)

func makeKey(password, salt string, iterations int) (string, error) {
	return string(pbkdf2.Key([]byte(password), []byte(salt), iterations, 256/8, sha256.New)[:]), nil
}

type register struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Kdf                int    `json:"kdf"`
	KdfIterations      int    `json:"kdfiterations"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var reg register
	len := r.ContentLength
	body := make([]byte, len)
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		fmt.Println("error when read request body:", err)
		return
	}
	err = json.Unmarshal(body, &reg)
	if err != nil {
		fmt.Println("error when unmarshal request body:", err)
		return
	}
	fmt.Println(reg)
}

func main() {
	server := http.Server{
		Addr: "127.0.0.1:4567",
	}
	http.HandleFunc("/api/accounts/register", handleRegister)
	server.ListenAndServe()
}
