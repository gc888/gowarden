package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"path"

	"regexp"

	"time"

	"github.com/404cn/gowarden/ds"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFileName   = "gowarden-db" // Database file name.
	accountTable = `CREATE TABLE IF NOT EXISTS "accounts" (
                        id TEXT,
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
	folderTable = `CREATE TABLE IF NOT EXISTS "folders" (
                        id TEXT,
                        name TEXT,
                        revisionDate INTEGER,
                        accountId TEXT,
                        PRIMARY KEY(id)
                    )`
	cipherTable = `CREATE TABLE IF NOT EXISTS "ciphers" (
                        id TEXT,
                        accountId TEXT,
                        revisionDate INTEGER,
                        type INTEGER,
                        folderId TEXT,
                        organizationId TEXT,
                        notes TEXT,
                        favorite INTEGER NOT NULL,
                        response TEXT,
                        username TEXT,
                        password TEXT,
                        passwordRevisionDate INTEGER,
                        totp TEXT,
                        PRIMARY KEY(id)
                   )`

	uriTable = `CREATE TABLE IF NOT EXISTS "uris" (
                    id TEXT,
                    cipherId TEXT,
                    response TEXT,
                    match TEXT,
                    uri TEXT,
                    PRIMARY KEY(id)
                )`
)

type DB struct {
	db  *sql.DB
	dir string
}

func New() *DB {
	return &DB{}
}

var StdDB = New()

func (db *DB) AddCipher(cipher *ds.Cipher, accountID string) (ds.Cipher, error) {
	return ds.Cipher{}, nil
}

func (db *DB) DeleteFolder(folderUUID string) error {
	stmt, err := db.db.Prepare("DELETE FROM folders WHERE id=$1")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(folderUUID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) RenameFolder(name, folderUUID string) (ds.Folder, error) {
	stmt, err := db.db.Prepare("UPDATE folders SET name=$1, revisionDate=$2 WHERE id=$3")
	if err != nil {
		return ds.Folder{}, err
	}

	tnow := time.Now()

	folder := ds.Folder{
		Id:           folderUUID,
		Name:         name,
		RevisionDate: tnow.UTC(),
		Object:       "folder",
	}

	_, err = stmt.Exec(name, tnow.Unix(), folderUUID)
	if err != nil {
		return ds.Folder{}, err
	}

	return folder, nil
}

func (db *DB) AddFolder(accountId, name string) (ds.Folder, error) {
	stmt, err := db.db.Prepare("INSERT INTO folders (id, name, revisionDate, accountId) VALUES(?, ?, ?, ?)")
	if err != nil {
		return ds.Folder{}, err
	}

	folderId := uuid.Must(uuid.NewRandom())

	folder := ds.Folder{
		Id:           folderId.String(),
		Name:         name,
		RevisionDate: time.Now(),
		Object:       "folder",
	}

	_, err = stmt.Exec(folderId, name, folder.RevisionDate.Unix(), accountId)
	if err != nil {
		return ds.Folder{}, nil
	}

	return folder, nil
}

func (db *DB) UpdateAccount(acc ds.Account) error {
	stmt, err := db.db.Prepare("UPDATE accounts SET refreshToken=$1, publicKey=$2, encryptedPrivateKey=$3 WHERE email=$4")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.RefreshToken, acc.Keys.PublicKey, acc.Keys.EncryptedPrivateKey, acc.Email)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetAccount(s string) (ds.Account, error) {
	var acc ds.Account
	acc.Keys = ds.Keys{}

	var validEmail = regexp.MustCompile(`(\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3})`)

	// TODO test
	if validEmail.MatchString(s) {
		// Get account from email.
		err := db.db.QueryRow("SELECT * FROM accounts WHERE email=?", s).Scan(&acc.Id, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.KdfIterations, &acc.Keys.PublicKey, &acc.Keys.EncryptedPrivateKey, &acc.RefreshToken)

		if err != nil {
			return acc, err
		}
	} else {
		// Get account from refresh token.
		err := db.db.QueryRow("SELECT * FROM accounts WHERE refreshToken=?", s).Scan(&acc.Id, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.KdfIterations, &acc.Keys.PublicKey, &acc.Keys.EncryptedPrivateKey, &acc.RefreshToken)

		if err != nil {
			return acc, err
		}
	}

	return acc, nil
}

func (db *DB) AddAccount(acc ds.Account) error {
	stmt, err := db.db.Prepare("INSERT INTO accounts(id, name, email, masterPasswordHash, masterPasswordHint, key, kdfIterations, publicKey, encryptedPrivateKey, refreshToken) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(uuid.Must(uuid.NewRandom()), acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, acc.KdfIterations, acc.Keys.PublicKey, acc.Keys.EncryptedPrivateKey, acc.RefreshToken)
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

	for _, sql := range []string{accountTable, folderTable, cipherTable} {
		if _, err := db.db.Exec(sql); err != nil {
			return errors.New(fmt.Sprintf("Sql error with %s\n%s", sql, err.Error()))
		}
	}
	return nil
}
