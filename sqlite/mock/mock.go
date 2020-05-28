package mock

import (
	"github.com/404cn/gowarden/ds"
)

type Mock struct{}

func New() *Mock {
	return &Mock{}
}

func (mock *Mock) DeleteAttachment(s1, s2 string) (string, error) {
	return "", nil
}

func (mock *Mock) AddAttachment(s string, att ds.Attachment) (ds.Cipher, error) {
	return ds.Cipher{}, nil
}

func (mock *Mock) GetFolders(s string) ([]ds.Folder, error) {
	return []ds.Folder{}, nil
}

func (mock *Mock) GetCiphers(s string) ([]ds.Cipher, error) {
	return []ds.Cipher{}, nil
}

func (mock *Mock) DeleteCipher(s1, s2 string) error {
	return nil
}

func (mock *Mock) UpdateCipher(cipher ds.Cipher, s string) (ds.Cipher, error) {
	return ds.Cipher{}, nil
}

func (mock *Mock) AddCipher(cipher ds.Cipher, s string) (ds.Cipher, error) {
	return ds.Cipher{}, nil
}

func (mock *Mock) RenameFolder(s1, s2 string) (ds.Folder, error) {
	return ds.Folder{}, nil
}

func (mock *Mock) DeleteFolder(s string) error {
	return nil
}

func (mock *Mock) GetAttachment(s1, s2 string) (ds.Attachment, error) {
	return ds.Attachment{}, nil
}

func (mock *Mock) AddFolder(accountId, name string) (ds.Folder, error) {
	return ds.Folder{}, nil
}

func (mock *Mock) AddAccount(acc ds.Account) error {
	return nil
}

func (mock *Mock) GetAccount(s string) (ds.Account, error) {
	return ds.Account{
		Email:         s,
		RefreshToken:  s,
		Kdf:           0,
		KdfIterations: 100000,
	}, nil
}

func (mock *Mock) UpdateAccount(acc ds.Account) error {
	return nil
}
