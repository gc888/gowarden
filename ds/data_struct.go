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
}

// Just for test, ready to delete.
func (acc *Account) ToString() {
	fmt.Printf("Name : %v\nEmail : %v\nMasterPasswordHash : %v\nMasterPasswordHint : %v\nKey : %v\nKdf : %v\nKdfIterations : %v\n", acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, acc.Kdf, acc.KdfIterations)
}
