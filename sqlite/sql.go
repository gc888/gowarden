package sqlite

import (
	"database/sql"
	"encoding/json"
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
                        favorite INTEGER NOT NULL,
                        data REAL,
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

func (db *DB) GetFolders(accId string) ([]ds.Folder, error) {
	var folders []ds.Folder

	rows, err := db.db.Query("SELECT id, name, revisionDate FROM folders WHERE accountId=?", accId)
	if err != nil {
		return folders, err
	}
	defer rows.Close()

	for rows.Next() {
		var folder ds.Folder
		var revData int64
		err = rows.Scan(&folder.Id, &folder.Name, &revData)
		if err != nil {
			return folders, err
		}

		folder.RevisionDate = time.Unix(revData, 0)

		folders = append(folders, folder)
	}

	if len(folders) < 1 {
		folders = make([]ds.Folder, 0)
	}
	return folders, err
}

func makeNewCipher(cipher *ds.Cipher) {
	cipher.Card = nil
	cipher.Fields = nil
	cipher.Identity = nil
	cipher.Name = cipher.Data.Name

	// Set ciph.Data.Uris if it's not in the DB
	if cipher.Data.Uri != nil && cipher.Data.Uris == nil {
		cipher.Data.Uris = []ds.Uri{ds.Uri{
			Uri:   cipher.Data.Uri,
			Match: nil,
		}}
	}

	if cipher.Data.Username != nil {
		cipher.Login = ds.Login{
			Username: cipher.Data.Username,
			Totp:     cipher.Data.Totp,
			Uri:      cipher.Data.Uri,
			Uris:     cipher.Data.Uris,
			Password: cipher.Data.Password,
		}
	}

	cipher.Notes = cipher.Data.Notes
	if cipher.Notes != nil {
		cipher.SecureNote = ds.SecureNote{
			Type: 0,
		}
	}
}

func row2cipher(row *sql.Rows) (ds.Cipher, error) {
	cipher := ds.Cipher{
		Favorite:            false,
		Edit:                true,
		OrganizationUseTotp: false,
		Object:              "cipher",
		Attachments:         nil,
		FolderId:            nil,
	}

	var accId string
	var favorite int
	var revDate int64
	var jsonData []byte
	var folderId sql.NullString
	err := row.Scan(&cipher.Id, &accId, &revDate, &cipher.Type, &folderId, &favorite, &jsonData)
	if err != nil {
		return cipher, err
	}

	// data 以外的字段手动构造
	err = json.Unmarshal(jsonData, &cipher.Data)
	if err != nil {
		return cipher, err
	}

	if favorite == 1 {
		cipher.Favorite = true
	}

	cipher.RevisionDate = time.Unix(revDate, 0)

	if folderId.Valid {
		cipher.FolderId = &folderId.String
	}

	makeNewCipher(&cipher)
	return cipher, nil
}

func (db *DB) GetCiphers(accId string) ([]ds.Cipher, error) {
	var ciphers []ds.Cipher

	rows, err := db.db.Query("SELECT * FROM ciphers WHERE accountId=$1", accId)
	if err != nil {
		return ciphers, err
	}

	for rows.Next() {
		var cipher ds.Cipher
		cipher, err = row2cipher(rows)
		if err != nil {
			return ciphers, err
		}

		ciphers = append(ciphers, cipher)
	}

	return ciphers, nil
}

func (db *DB) DeleteCipher(accId, cipherId string) error {
	stmt, err := db.db.Prepare("DELETE FROM ciphers WHERE id=$1 AND accountId=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(cipherId, accId)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateCipher(cipher ds.Cipher, accId string, cipherId string) error {
	favorite := 0
	if cipher.Favorite {
		favorite = 1
	}

	stmt, err := db.db.Prepare("UPDATE ciphers SET type=$1, revisionDate=$2, data=$3, folderId=$4, favorite=$5 WHERE id=$6 and accountId=$7")
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(&cipher.Data)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(cipher.Type, time.Now().Unix(), jsonData, cipher.FolderId, favorite, cipherId, accId)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AddCipher(cipher ds.Cipher, accId string) (ds.Cipher, error) {
	cipher.RevisionDate = time.Now()

	stmt, err := db.db.Prepare("INSERT INTO ciphers values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, err
	}

	// 字段太多，直接jsonify存起来
	b, err := json.Marshal(&cipher.Data)
	if err != nil {
		return cipher, nil
	}

	cipherId := uuid.Must(uuid.NewRandom())

	// TODO
	_, err = stmt.Exec(cipherId, accId, cipher.RevisionDate.Unix(), cipher.Type, cipher.FolderId, 0, b)
	if err != nil {
		return cipher, err
	}

	makeNewCipher(&cipher)
	return cipher, nil
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
