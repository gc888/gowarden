package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"path"

	"github.com/404cn/gowarden/ds"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFileName   = "gowarden-db" // Database file name.
	accountTable = `CREATE TABLE IF NOT EXISTS "accounts" (
                        id INTEGER,
                        name TEXT,
                        email TEXT UNIQUE,
                        masterPasswordHash TEXT,
                        masterPasswordHint TEXT,
                        key INTEGER,
                        kdfIterations INTEGER,
                        publicKey TEXT NOT NULL,
                        encryptedPrivateKey TEXT NOT NULL,
                        refreshToken TEXT,
                        PRIMARY KEY(id)
                    )`  // User's account table
)

type DB struct {
	db  *sql.DB
	dir string
}

func New() *DB {
	return &DB{}
}

var StdDB = New()

func (db *DB) UpdateAccount(acc ds.Account) error {
	stmt, err := db.db.Prepare("UPDATE accounts set refreshToken=$1 WHERE email=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.RefreshToken, acc.Email)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetAccount(email string) (ds.Account, error) {
	var acc ds.Account
	acc.Keys = ds.Keys{}

	var id int
	err := db.db.QueryRow("SELECT * FROM accounts WHERE email=?", email).Scan(&id, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.KdfIterations, &acc.Keys.PublicKey, &acc.Keys.EncryptedPrivateKey, &acc.RefreshToken)
	if err != nil {
		return acc, err
	}

	return acc, nil
}

func (db *DB) AddAccount(acc ds.Account) error {
	stmt, err := db.db.Prepare("INSERT INTO accounts(name, email, masterPasswordHash, masterPasswordHint, key, kdfIterations, publicKey, encryptedPrivateKey, refreshToken) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, acc.KdfIterations, acc.Keys.PublicKey, acc.Keys.EncryptedPrivateKey, acc.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Open() error {
	var err error
	if db.dir != "" {
		db.db, err = sql.Open("sqlite3", path.Join(db.dir, dbFileName))
	} else {
		db.db, err = sql.Open("sqlite3", dbFileName)
	}
	return err
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) SetDir(d string) {
	db.dir = d
}

func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (db *DB) Init() error {
	if PathExist(dbFileName) {
		err := os.Remove(dbFileName)
		if err != nil {
			return err
		}
	}

	for _, sql := range []string{accountTable} {
		if _, err := db.db.Exec(sql); err != nil {
			return errors.New(fmt.Sprintf("Sql error with %s\n%s", sql, err.Error()))
		}
	}
	return nil
}
