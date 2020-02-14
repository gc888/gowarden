package ds

type Login struct {
	Grant_type       string `json:"grant_type"`
	UserName         string `json:"userName"`
	Password         string `json:"password"`
	Scope            string `json:"scope"`
	Client_id        string `json:"client_id"`
	DeviceType       int    `json:"deviceType"`
	DeviceIdentifier string `json:"deviceIdentifier"`
	DeviceName       string `json:"deviceName"`
}

type Account struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Kdf                int    `json:"kdf"`
	KdfIterations      int    `json:"kdfiterations"`
	Keys               Keys   `json:"keys"`
}

type Keys struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}
