package ds

import (
	"time"
)

// type to handle folders's response
type Folder struct {
	Id           string    `json:"Id"`
	Name         string    `json:"Name"`
	RevisionDate time.Time `json:"RevisionDate"`
	Object       string    `json:"Object"`
}

type Cipher struct {
	Type           int      `json:"type"`
	FolderId       string   `json:"folderId"`
	OrganizationId string   `json:"organizationId"`
	Name           string   `json:"name"`
	Notes          string   `json:"Name"`
	Favorite       bool     `json:"favorite"`
	Login          Login    `json:"login"`
	Fields         []Fields `json:"fields"`
}

type Login struct {
	Response             string `json:"response"`
	Uris                 []Uris `json:"uris"`
	UserName             string `json:"username"`
	Password             string `json:"password"`
	PasswordRevisionDate string `json:"passwordRevisionDate"`
	Totp                 string `json:"totp"`
}

type Uris struct {
	Response string `json:"response"`
	Match    string `json:"match"`
	Uri      string `json:"uri"`
}

type Fields struct {
	Response string `json:"response"`
	Type     int    `josn:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

type Account struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Kdf                int    `json:"kdf"`
	KdfIterations      int    `json:"kdfiterations"`
	Keys               Keys   `json:"keys"`
	RefreshToken       string `json:"refresh_token"`
}

type Keys struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}
