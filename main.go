package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"golang.org/x/crypto/pbkdf2"
)

func makeKey(password, salt string, iterations int) (string, error) {
	return string(pbkdf2.Key([]byte(password), []byte(salt), iterations, 256/8, sha256.New)[:]), nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, r)
}

func main() {
	server := http.Server{
		Addr: "127.0.0.1:4567",
	}
	http.HandleFunc("/accounts/register", handleRegister)
	server.ListenAndServe()
}
