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
                        favorite INTEGER NOT NULL,
						name TEXT,
						notes TEXT,
                        PRIMARY KEY(id)
                    )`
	// TODO how to handle totp
	loginTable = `CREATE TABLE IF NOT EXISTS "logins" (
                        id TEXT,
                        cipherId TEXT,
						username TEXT,
						password TEXT,
 						totp INTEGER,
                        PRIMARY KEY(id)
                    )`

	uriTable = `CREATE TABLE IF NOT EXISTS "uris" (
                        id TEXT,
                        cipherId TEXT,
						match INTEGER,
						uri TEXT,
                        PRIMARY KEY(id)
                    )`
	fieldTable = `CREATE TABLE IF NOT EXISTS "fields" (
                        id TEXT,
                        cipherId TEXT,
						type INTEGER,
						name TEXT,
						value TEXT,
                        PRIMARY KEY(id)
                    )`
)

// FIXME uri and fields in databse is empty

type DB struct {
	db  *sql.DB
	dir string
}

func New() *DB {
	return &DB{}
}

func (db *DB) AddCipher(cipher ds.Cipher, accId string) (ds.Cipher, error) {
	cipher.Id = uuid.Must(uuid.NewRandom()).String()
	cipher.RevisionDate = time.Now()
	var favorite int

	if cipher.Favorite {
		favorite = 1
	}

	cipherStmt, err := db.db.Prepare("INSERT INTO ciphers VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, nil
	}

	loginStmt, err := db.db.Prepare("INSERT INTO logins VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, nil
	}

	uriStmt, err := db.db.Prepare("INSERT INTO uris VALUES(?, ?, ?, ?)")
	if err != nil {
		return cipher, nil
	}

	fieldStmt, err := db.db.Prepare("INSERT INTO fields VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, nil
	}

	_, err = cipherStmt.Exec(cipher.Id, accId, cipher.RevisionDate.Unix(), cipher.Type, cipher.FolderId, favorite, cipher.Name, cipher.Notes)
	if err != nil {
		return cipher, nil
	}

	// TODO totp's handle
	_, err = loginStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, cipher.Login.Username, cipher.Login.Password, 0)
	if err != nil {
		return cipher, nil
	}

	for _, uri := range cipher.Login.Uris {
		_, err = uriStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, uri.Match, uri.Uri)
		if err != nil {
			return cipher, nil
		}
	}

	for _, field := range cipher.Fields {
		_, err = fieldStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, field.Type, field.Name, field.Value)
		if err != nil {
			return cipher, nil
		}
	}

	makeNewCipher(&cipher)
	return cipher, nil
}

func makeNewCipher(cipher *ds.Cipher) {
	cipher.Object = "cipher"
	cipher.Edit = true

	if cipher.Login.Uris != nil {
		cipher.Login.Uri, cipher.Data.Uri = cipher.Login.Uris[0].Uri, cipher.Login.Uris[0].Uri
	}

	// 只有object为login时
	if cipher.Login.Username != nil {
		cipher.Data = ds.CipherData{
			Username: cipher.Login.Username,
			Password: cipher.Login.Password,
			Totp:     cipher.Login.Totp,
			Name:     cipher.Name,
			Notes:    cipher.Notes,
			Fields:   cipher.Fields,
			Uris:     cipher.Login.Uris,
		}
	}

}

func (db *DB) DeleteCipher(accId, cipherId string) error {
	cipherStmt, err := db.db.Prepare("DELETE FROM ciphers WHERE id=$1 AND accountId=$2")
	if err != nil {
		return err
	}

	_, err = cipherStmt.Exec(cipherId, accId)
	if err != nil {
		return err
	}

	loginStmt, err := db.db.Prepare("DELETE FROM logins WHERE cipherId=$1")
	if err != nil {
		return err
	}

	_, err = loginStmt.Exec(cipherId)
	if err != nil {
		return err
	}

	uriStmt, err := db.db.Prepare("DELETE FROM uris WHERE cipherId=$1")
	if err != nil {
		return err
	}

	_, err = uriStmt.Exec(cipherId)
	if err != nil {
		return err
	}

	fieldStmt, err := db.db.Prepare("DELETE FROM fields WHERE cipherId=$1")
	if err != nil {
		return err
	}

	_, err = fieldStmt.Exec(cipherId)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateCipher(cipher ds.Cipher, accId string) error {
	now := time.Now()
	favorite := 0
	if cipher.Favorite {
		favorite = 1
	}

	cipherStmt, err := db.db.Prepare("UPDATE ciphers SET revisionDate=$1, type=$2, folderId=$3, favorite=$4, name=$5, notes=$6 WHERE id=$7 AND accountId=$8")
	if err != nil {
		return err
	}

	_, err = cipherStmt.Exec(now.Unix(), cipher.Type, cipher.FolderId, favorite, cipher.Name, cipher.Notes, cipher.Id, accId)
	if err != nil {
		return err
	}

	loginStmt, err := db.db.Prepare("UPDATE logins SET username=$1, password=$2, totp=$3 WHERE cipherId=$4")
	if err != nil {
		return err
	}

	_, err = loginStmt.Exec(cipher.Login.Username, cipher.Login.Password, cipher.Login.Totp, cipher.Id)
	if err != nil {
		return err
	}

	_, err = db.db.Exec("DELETE FROM uris WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return err
	}
	uriStmt, err := db.db.Prepare("INSERT INTO uris VALUES (?, ?, ? ,?)")
	if err != nil {
		return err
	}

	for _, uri := range cipher.Login.Uris {
		_, err = uriStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, uri.Match, uri.Uri)
		if err != nil {
			return err
		}
	}

	_, err = db.db.Exec("DELETE FROM fields WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return err
	}
	fieldStmt, err := db.db.Prepare("INSERT INTO fields VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	for _, field := range cipher.Fields {
		_, err = fieldStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, field.Type, field.Name, field.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetCiphers(accId string) ([]ds.Cipher, error) {
	var ciphers []ds.Cipher
	var cipherId string
	// 占位使用
	var fooId string

	db.db.QueryRow("SELECT DISTINCT id FROM ciphers WHERE accountId=$1", accId).Scan(&cipherId)

	cipherRows, err := db.db.Query("SELECT * FROM ciphers WHERE accountId=$1", accId)
	if err != nil {
		return ciphers, err
	}

	for cipherRows.Next() {
		var cipher ds.Cipher
		var revDate int64
		var favorite int

		err = cipherRows.Scan(&cipher.Id, &fooId, &revDate, &cipher.Type, &cipher.FolderId, &favorite, &cipher.Name, &cipher.Notes)
		if err != nil {
			return ciphers, err
		}

		if favorite == 1 {
			cipher.Favorite = true
		}

		cipher.RevisionDate = time.Unix(revDate, 0)

		ciphers = append(ciphers, cipher)
	}

	for i, cipher := range ciphers {
		loginRow := db.db.QueryRow("SELECT username, password, totp FROM logins WHERE cipherId=$1", cipher.Id)

		err = loginRow.Scan(&cipher.Login.Username, &cipher.Login.Password, &cipher.Login.Totp)
		if err != nil {
			return ciphers, err
		}

		uriRows, err := db.db.Query("SELECT match, uri FROM uris WHERE cipherId=$1", cipher.Id)
		if err != nil {
			return ciphers, err
		}
		var uris []ds.Uri
		for uriRows.Next() {
			var uri ds.Uri
			err := uriRows.Scan(&uri.Match, &uri.Uri)
			if err != nil {
				return ciphers, err
			}
			uris = append(uris, uri)
		}
		cipher.Login.Uris = uris

		fieldRows, err := db.db.Query("SELECT type, name, value FROM fields WHERE cipherId=$1", cipher.Id)
		if err != nil {
			return ciphers, err
		}

		var fields []ds.Field
		for fieldRows.Next() {
			var field ds.Field

			err = fieldRows.Scan(&field.Type, &field.Name, &field.Value)
			if err != nil {
				return ciphers, err
			}

			fields = append(fields, field)
		}

		cipher.Fields = fields

		makeNewCipher(&cipher)

		ciphers[i] = cipher
	}

	return ciphers, nil
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

	for _, sql := range []string{accountTable, folderTable, cipherTable, loginTable, uriTable, fieldTable} {
		if _, err := db.db.Exec(sql); err != nil {
			return errors.New(fmt.Sprintf("Sql error with %s\n%s", sql, err.Error()))
		}
	}
	return nil
}
