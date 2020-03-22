package mock

import (
	"github.com/404cn/gowarden/ds"
)

type Mock struct{}

func New() *Mock {
	return &Mock{}
}

func (mock *Mock) DeleteFolder(name, folderUUID string) (ds.Folder, error) {
	return ds.Folder{}, nil
}

func (mock *Mock) AddFolder(accountId, name string) (ds.Folder, error) {
	return ds.Folder{}, nil
}

func (mock *Mock) AddAccount(acc ds.Account) error {
	return nil
}

func (mock *Mock) GetAccount(s string) (ds.Account, error) {
	return ds.Account{
		Kdf:           0,
		KdfIterations: 100000,
	}, nil
}

func (mock *Mock) UpdateAccount(acc ds.Account) error {
	return nil
}
