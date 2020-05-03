package ds

import (
	"time"
)

// struct used in sync
type SyncData struct {
	Profile Profile
	Folders []Folder
	Ciphers []Cipher
	Domains Domains
	Object  string
}

type Domains struct {
	EquivalentDomains       []string
	GlobalEquivalentDomains []GlobalEquivalentDomains
	Object                  string
}

type GlobalEquivalentDomains struct {
	Type     int
	Domains  []string
	Excluded bool
}

// profile to in syncing
type Profile struct {
	Id                 string
	Name               *string
	Email              string
	EmailVerified      bool
	Premium            bool
	MasterPasswordHint string
	Culture            string
	TwoFactorEnabled   bool
	Key                string
	PrivateKey         string
	SecurityStamp      *string
	Organizations      []string
	Object             string
}

func (acc Account) Profile() Profile {
	p := Profile{
		Id:                 acc.Id,
		Name:               nil,
		Email:              acc.Email,
		EmailVerified:      false,
		Premium:            false,
		MasterPasswordHint: acc.MasterPasswordHint,
		Culture:            "en-US",
		TwoFactorEnabled:   false,
		Key:                acc.Key,
		PrivateKey:         acc.Keys.EncryptedPrivateKey,
		SecurityStamp:      nil,
		Organizations:      make([]string, 0),
		Object:             "profile",
	}

	return p
}

// structs about cipher
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
	Fields     []FieldsType
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
	Fields   []FieldsType
	Uris     []Uri
}

type FieldsType struct {
	Type  int
	Name  string
	Value string
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
