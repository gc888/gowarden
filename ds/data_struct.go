package ds

import "fmt"

type Account struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Kdf                int    `json:"kdf"`
	KdfIterations      int    `json:"kdfiterations"`
	Keys               keys   `json:"keys"`
}

type keys struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}

// Just for test, ready to delete.
func (acc *Account) ToString() {
	// TODO
	fmt.Printf("Name : %v\nEmail : %v\nMasterPasswordHash : %v\nMasterPasswordHint : %v\nKey : %v\nKdf : %v\nKdfIterations : %v\npublicKey:%v\nencryptedPrivateKey:%s\n", acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, acc.Kdf, acc.KdfIterations, acc.Keys.PublicKey, acc.Keys.EncryptedPrivateKey)
}
