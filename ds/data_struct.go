package ds

import (
	"time"
)

type Cipher struct {
	Type                int
	FolderId            *string
	OrganizationId      *string
	Favorite            bool
	Edit                bool
	Id                  string
	Data                CipherData
	Attachments         []string
	OrganizationUseTotp bool
	RevisionDate        time.Time
	Object              string
	CollectionIds       []string

	Card       *string
	Fields     []string
	Identity   *string
	Login      Login
	Name       *string
	Notes      *string
	SecureNote SecureNote
}

type CipherData struct {
	Uri      *string
	Username *string
	Password *string
	Totp     *string
	Name     *string
	Notes    *string
	Fields   []string
	Uris     []Uri
}

type Uri struct {
	Uri   *string
	Match *int
}

type Login struct {
	Password *string
	Totp     *string
	Uri      *string
	Uris     []Uri
	Username *string
}

type SecureNote struct {
	Type int
}

// type to handle folders's response
type Folder struct {
	Id           string
	Name         string
	RevisionDate time.Time
	Object       string
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
