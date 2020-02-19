package mock

import (
	"log"

	"github.com/404cn/gowarden/ds"
)

type Mock struct{}

func New() *Mock {
	return &Mock{}
}

func (mock *Mock) AddAccount(acc ds.Account) error {
	log.Println("mock add account")
	return nil
}

func (mock *Mock) GetAccount(s string) (ds.Account, error) {
	return ds.Account{}, nil
}
