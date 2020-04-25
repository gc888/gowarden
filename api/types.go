package api

import (
	"github.com/404cn/gowarden/ds"
	"go.uber.org/zap"
)

const (
	jwtExpiresin = 3600
)

type handler interface {
	AddAccount(ds.Account) error
	GetAccount(string) (ds.Account, error)
	UpdateAccount(ds.Account) error
	AddFolder(string, string) (ds.Folder, error)
	DeleteFolder(string) error
	RenameFolder(string, string) (ds.Folder, error)
	AddCipher(ds.Cipher, string) (ds.Cipher, error)
	UpdateCipher(ds.Cipher, string, string) error
	DeleteCipher(string, string) error
}

type APIHandler struct {
	db         handler
	signingKey string
	logger     *zap.SugaredLogger
}

func New(db handler, key string, sugar *zap.SugaredLogger) *APIHandler {
	return &APIHandler{
		db:         db,
		signingKey: key,
		logger:     sugar,
	}
}
