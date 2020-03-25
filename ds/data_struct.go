package ds

import (
	"time"
)

type Cipher struct {
	Object              string
	FolderId            string
	Favorite            bool
	Edit                bool
	Id                  string
	OrganizationId      string
	Type                int
	Data                CipherData
	Name                string
	Notes               string
	Login               Login
	Card                string
	Identiey            string
	SecureNote          SecureNote
	Fields              []Fields
	PasswordHistory     string
	Attachments         []string
	OrganizationUseTotp bool
	RevisionDate        time.Time
}

type SecureNote struct {
	Type int
}

type CipherData struct {
	Uri                  string
	Uris                 []Uris
	Username             string
	Password             string
	PasswordRevisionDate time.Time
	Totp                 string
	Name                 string
	Notes                string
	Fields               string
	PasswordHistory      string
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
	Match string `json:"match"`
	Uri   string `json:"uri"`
}

type Fields struct {
	Response string `json:"response"`
	Type     int    `josn:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

// type to handle folders's response
type Folder struct {
	Id           string    `json:"Id"`
	Name         string    `json:"Name"`
	RevisionDate time.Time `json:"RevisionDate"`
	Object       string    `json:"Object"`
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
