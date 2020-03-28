package ds

import (
	"time"
)

type Cipher struct {
	// 类型, 一般为cipher
	Object string
	// 与之关联的文件夹id, 没有则为nil
	FolderId string
	// 是否为favorite
	Favorite bool
	// TODO 暂时为true, 未找到false的情况
	Edit bool
	// 表示一个唯一cipher的id
	Id string
	// TODO 组织id, 暂时不知道作用
	OrganizationId string
	// 类型, 详见API.org
	Type  int
	Data  CipherData
	Name  string
	Notes string

	// four type of cipher
	Login      Login
	Card       string
	Identiey   []string
	SecureNote SecureNote

	Fields []Fields
	// TODO 暂时为nil
	PasswordHistory string
	// TODO 附件
	Attachments []string
	// TODO some totp thing about origianization
	OrganizationUseTotp bool
	// 时间 例如:2020-03-25T15:05:32.26576882
	RevisionDate time.Time
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
	Fields               []Fields
	PasswordHistory      string
}

type Login struct {
	Uri                  string
	Uris                 []Uris
	UserName             string
	Password             string
	PasswordRevisionDate string
	Totp                 string
}

type Uris struct {
	Match string
	Uri   string
}

type Fields struct {
	Type  int
	Name  string
	Value string
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
