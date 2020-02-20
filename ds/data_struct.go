package ds

type Account struct {
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
