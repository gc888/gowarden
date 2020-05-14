package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
)

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// getIV get initialization vector
func getIV() ([]byte, error) {
	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return iv, err
	}
	return iv, nil
}

func cipherString(tp, iv, ct, mac string) string {
	return tp + "." + iv + "|" + ct + "|" + mac
}

// pt: A secret note
// key: enckey top 32 byte of master key
// mackey: mackey count from iv + cipher text
func Encrypt(pt string, key, macKey []byte) string {
	iv, err := getIV()
	if err != nil {
		panic(err)
	}

	ct := aes256(pt, key, iv)
	h := hmac.New(sha256.New, macKey)
	h.Write([]byte(ct + string(iv)))
	mac := h.Sum(nil)

	cs := cipherString("2", base64.StdEncoding.EncodeToString(iv), base64.StdEncoding.EncodeToString([]byte(ct)), base64.StdEncoding.EncodeToString(mac))

	return cs
}

func aes256(plaintext string, key, iv []byte) string {
	bPlaintext := pKCS5Padding([]byte(plaintext), 32, len(plaintext))
	block, _ := aes.NewCipher(key)
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, bPlaintext)
	return hex.EncodeToString(ciphertext)
}

func pKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
