package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/404cn/gowarden/utils"
	"os"
	"path"
	"strconv"

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
	loginTable = `CREATE TABLE IF NOT EXISTS "logins" (
                        id TEXT,
                        cipherId TEXT,
						username TEXT,
						password TEXT,
 						totp TEXT,
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
	attachmentTable = `CREATE TABLE IF NOT EXISTS "attachments" (
                        id TEXT,
                        cipherId TEXT,
						filename TEXT,
						key Text,
						size Text,
						url TEXT,
                        PRIMARY KEY(id)
                    )`
	cardTable = `CREATE TABLE IF NOT EXISTS "cards" (
                        id TEXT,
                        cipherId TEXT,
						cardholdername TEXT,
						brand TEXT,
						number TEXT,
						expmonth TEXT,
						expyear TEXT,
						code TEXT,
                        PRIMARY KEY(id)
                    )`
	identityTable = `CREATE TABLE IF NOT EXISTS "identities" (
                        id TEXT,
                        cipherId TEXT,
						title TEXT,
						firstname TEXT,
						middlename TEXT,
						lastname TEXT,
						address1 TEXT,
						address2 TEXT,
						address3 TEXT,
						city TEXT,
						state TEXT,
						postalcode TEXT,
						country TEXT,
						company TEXT,
						email TEXT,
						phone TEXT,
						ssn TEXT,
						username TEXT,
						passportnumber TEXT,
						licensenumber TEXT,
                        PRIMARY KEY(id)
                    )`
)

// FIXME ssn, cardholdername didn't save

type DB struct {
	db  *sql.DB
	dir string
}

func New() *DB {
	return &DB{}
}

// TODO
func (db *DB) SaveCSV(csvs []ds.CSV) error {

	return nil
}

func (db *DB) AddAttachment(cipherId string, attachment ds.Attachment) (ds.Cipher, error) {
	cipher, err := getCipher(db.db, cipherId)
	if err != nil {
		return cipher, err
	}

	attachment.Object = "attachment"
	size, err := strconv.Atoi(attachment.Size)
	if err != nil {
		return cipher, err
	}
	attachment.SizeName = strconv.FormatInt(int64(size>>10), 10) + " KB"

	stmt, err := db.db.Prepare("INSERT INTO attachments VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, err
	}

	_, err = stmt.Exec(attachment.Id, cipherId, attachment.FileName, attachment.Key, attachment.Size, attachment.Url)
	if err != nil {
		return cipher, err
	}

	cipher.Attachments = append(cipher.Attachments, attachment)

	return cipher, nil
}

func (db *DB) DeleteAttachment(cipherId, attachmentId string) (url string, err error) {
	url, err = getAttachmentUrl(db.db, cipherId, attachmentId)
	stmt, err := db.db.Prepare("DELETE FROM attachments WHERE id=$1 AND cipherID=$2")
	if err != nil {
		return "", err
	}

	_, err = stmt.Exec(attachmentId, cipherId)
	if err != nil {
		return "", err
	}

	return url, nil
}

func getAttachmentUrl(db *sql.DB, cipherId, attachmentId string) (url string, err error) {
	err = db.QueryRow("SELECT url FROM attachments WHERE id=$1 AND cipherId=$2", attachmentId, cipherId).Scan(&url)
	return url, err
}

func (db *DB) GetAttachment(cipherId, attachmentId string) (ds.Attachment, error) {
	var attachment ds.Attachment

	err := db.db.QueryRow("SELECT id, filename, key, size, url FROM attachments WHERE cipherId=$1", cipherId).Scan(
		&attachment.Id, &attachment.FileName, &attachment.Key, &attachment.Size, &attachment.Url)
	if err != nil {
		return attachment, err
	}

	attachment.Object = "attachment"
	size, err := strconv.Atoi(attachment.Size)
	if err != nil {
		return attachment, err
	}
	attachment.SizeName = strconv.FormatInt(int64(size>>10), 10) + " KB"

	return attachment, nil
}

func getAttachments(db *DB, cipherId string) ([]ds.Attachment, error) {
	var attachments []ds.Attachment

	rows, err := db.db.Query("SELECT id, filename, key, size, url FROM attachments WHERE cipherId=$1", cipherId)
	if err != nil {
		return attachments, err
	}

	for rows.Next() {
		var attachment ds.Attachment

		err = rows.Scan(&attachment.Id, &attachment.FileName, &attachment.Key, &attachment.Size, &attachment.Url)
		if err != nil {
			return attachments, err
		}

		makeNewAttachment(&attachment)

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func makeNewAttachment(attachment *ds.Attachment) {
	attachment.Object = "attachment"
	size, _ := strconv.Atoi(attachment.Size)
	attachment.SizeName = strconv.FormatInt(int64(size>>10), 10) + " KB"
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

	cardStmt, err := db.db.Prepare("INSERT INTO cards VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, err
	}
	_, err = cardStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, cipher.Card.CardHolderName, cipher.Card.Brand, cipher.Card.Number, cipher.Card.ExpMonth, cipher.Card.ExpYear, cipher.Card.Code)
	if err != nil {
		return cipher, err
	}

	identityStmt, err := db.db.Prepare("INSERT INTO identities VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ,? ,? ,?,?,?,?,?,?)")
	if err != nil {
		return cipher, err
	}
	_, err = identityStmt.Exec(
		uuid.Must(uuid.NewRandom()).String(),
		cipher.Id,
		cipher.Identity.Title,
		cipher.Identity.FirstName,
		cipher.Identity.MiddleName,
		cipher.Identity.LastName,
		cipher.Identity.Address1,
		cipher.Identity.Address2,
		cipher.Identity.Address3,
		cipher.Identity.City,
		cipher.Identity.State,
		cipher.Identity.PostalCode,
		cipher.Identity.Country,
		cipher.Identity.Company,
		cipher.Identity.Email,
		cipher.Identity.Phone,
		cipher.Identity.Ssn,
		cipher.Identity.Username,
		cipher.Identity.PassportNumber,
		cipher.Identity.LicenseNumber,
	)
	if err != nil {
		return cipher, err
	}

	makeNewCipher(&cipher)
	return cipher, nil
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

	attachmentStmt, err := db.db.Prepare("DELETE FROM attachments WHERE cipherId=$1")
	if err != nil {
		return err
	}
	_, err = attachmentStmt.Exec(cipherId)
	if err != nil {
		return err
	}

	cardStmt, err := db.db.Prepare("DELETE FROM cards WHERE cipherId=$1")
	if err != nil {
		return err
	}
	_, err = cardStmt.Exec(cipherId)
	if err != nil {
		return err
	}

	_, err = db.db.Exec("DELETE FROM identities WHERE cipherId=$1", cipherId)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateCipher(cipher ds.Cipher, accId string) (ds.Cipher, error) {
	cipher.RevisionDate = time.Now()
	favorite := 0
	if cipher.Favorite {
		favorite = 1
	}

	cipherStmt, err := db.db.Prepare("UPDATE ciphers SET revisionDate=$1, type=$2, folderId=$3, favorite=$4, name=$5, notes=$6 WHERE id=$7 AND accountId=$8")
	if err != nil {
		return cipher, err
	}

	_, err = cipherStmt.Exec(cipher.RevisionDate.Unix(), cipher.Type, cipher.FolderId, favorite, cipher.Name, cipher.Notes, cipher.Id, accId)
	if err != nil {
		return cipher, err
	}

	loginStmt, err := db.db.Prepare("UPDATE logins SET username=$1, password=$2, totp=$3 WHERE cipherId=$4")
	if err != nil {
		return cipher, err
	}

	_, err = loginStmt.Exec(cipher.Login.Username, cipher.Login.Password, cipher.Login.Totp, cipher.Id)
	if err != nil {
		return cipher, err
	}

	_, err = db.db.Exec("DELETE FROM uris WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	uriStmt, err := db.db.Prepare("INSERT INTO uris VALUES (?, ?, ? ,?)")
	if err != nil {
		return cipher, err
	}

	for _, uri := range cipher.Login.Uris {
		_, err = uriStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, uri.Match, uri.Uri)
		if err != nil {
			return cipher, err
		}
	}

	_, err = db.db.Exec("DELETE FROM fields WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	fieldStmt, err := db.db.Prepare("INSERT INTO fields VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return cipher, err
	}

	for _, field := range cipher.Fields {
		_, err = fieldStmt.Exec(uuid.Must(uuid.NewRandom()).String(), cipher.Id, field.Type, field.Name, field.Value)
		if err != nil {
			return cipher, err
		}
	}

	cipher.Attachments, err = getAttachments(db, cipher.Id)
	if err != nil {
		return cipher, err
	}

	// FIXME 更新后数据没了
	cardStmt, err := db.db.Prepare("UPDATE cards SET cardholdername=$1, brand=$2, number=$3, expmonth=$4, expyear=$5, code=$6 WHERE cipherId=$7")
	if err != nil {
		return cipher, err
	}
	_, err = cardStmt.Exec(cipher.Card.CardHolderName, cipher.Card.Brand, cipher.Card.Number, cipher.Card.ExpMonth, cipher.Card.ExpYear, cipher.Card.Code, cipher.Id)
	if err != nil {
		return cipher, err
	}

	// FIXME
	_, err = db.db.Exec("UPDATE identities SET title=$1, firstname=$2, middlename=$3, lastname=$4, address1=$5, address2=$6, address3=$7, city=$8, state=$9, postalcode=$10, country=$11, company=$12, email=$13, phone=$14, ssn=$15, username=$16, passportnumber=$17, licensenumber=$18 WHERE cipherId=$19",
		cipher.Identity.Title,
		cipher.Identity.FirstName,
		cipher.Identity.MiddleName,
		cipher.Identity.LastName,
		cipher.Identity.Address1,
		cipher.Identity.Address2,
		cipher.Identity.Address3,
		cipher.Identity.City,
		cipher.Identity.State,
		cipher.Identity.PostalCode,
		cipher.Identity.Country,
		cipher.Identity.Company,
		cipher.Identity.Email,
		cipher.Identity.Phone,
		cipher.Identity.Ssn,
		cipher.Identity.Username,
		cipher.Identity.PassportNumber,
		cipher.Identity.LicenseNumber,
		cipher.Id)
	if err != nil {
		return cipher, err
	}

	return cipher, nil
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

		attachmentRows, err := db.db.Query("SELECT id, filename, key, size, url FROM attachments WHERE cipherId=$1", cipher.Id)
		if err != nil {
			return ciphers, err
		}
		var attachments []ds.Attachment
		for attachmentRows.Next() {
			var attachment ds.Attachment
			err = attachmentRows.Scan(&attachment.Id, &attachment.FileName, &attachment.Key, &attachment.Size, &attachment.Url)
			if err != nil {
				return ciphers, err
			}
			attachment.Object = "attachment"
			size, err := strconv.Atoi(attachment.Size)
			if err != nil {
				return ciphers, err
			}
			attachment.SizeName = strconv.FormatInt(int64(size>>10), 10) + " KB"
			attachments = append(attachments, attachment)
		}
		cipher.Attachments = attachments

		cardRow := db.db.QueryRow("SELECT cardholdername, brand, number, expmonth, expyear, code FROM cards WHERE cipherId=$1", cipher.Id)
		err = cardRow.Scan(&cipher.Card.CardHolderName, &cipher.Card.Brand, &cipher.Card.Number, &cipher.Card.ExpMonth, &cipher.Card.ExpYear, &cipher.Card.Code)
		if err != nil {
			return ciphers, err
		}

		identityRow := db.db.QueryRow("SELECT * FROM identities WHERE cipherId=$1", cipher.Id)
		var foo, bar string
		err = identityRow.Scan(
			&foo,
			&bar,
			&cipher.Identity.Title,
			&cipher.Identity.FirstName,
			&cipher.Identity.MiddleName,
			&cipher.Identity.LastName,
			&cipher.Identity.Address1,
			&cipher.Identity.Address2,
			&cipher.Identity.Address3,
			&cipher.Identity.City,
			&cipher.Identity.State,
			&cipher.Identity.PostalCode,
			&cipher.Identity.Country,
			&cipher.Identity.Company,
			&cipher.Identity.Email,
			&cipher.Identity.Phone,
			&cipher.Identity.Ssn,
			&cipher.Identity.Username,
			&cipher.Identity.PassportNumber,
			&cipher.Identity.LicenseNumber)
		if err != nil {
			return ciphers, err
		}

		makeNewCipher(&cipher)

		ciphers[i] = cipher
	}

	return ciphers, nil
}

func getCipher(db *sql.DB, cipherId string) (ds.Cipher, error) {
	var cipher ds.Cipher
	var revDate int64
	var favorite int

	cipher.Id = cipherId

	db.QueryRow("SELECT revisionDate, type, folderId, favorite, name, notes FROM ciphere WHERE id=$1", cipherId).Scan(
		&revDate, &cipher.Type, &cipher.FolderId, &favorite, &cipher.Name, &cipher.Notes)

	cipher.RevisionDate = time.Unix(revDate, 0)
	if favorite == 1 {
		cipher.Favorite = true
	}

	loginRow := db.QueryRow("SELECT username, password, totp FROM logins WHERE cipherId=$1", cipher.Id)

	err := loginRow.Scan(&cipher.Login.Username, &cipher.Login.Password, &cipher.Login.Totp)
	if err != nil {
		return cipher, err
	}

	uriRows, err := db.Query("SELECT match, uri FROM uris WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	var uris []ds.Uri
	for uriRows.Next() {
		var uri ds.Uri
		err := uriRows.Scan(&uri.Match, &uri.Uri)
		if err != nil {
			return cipher, err
		}
		uris = append(uris, uri)
	}
	cipher.Login.Uris = uris

	fieldRows, err := db.Query("SELECT type, name, value FROM fields WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}

	var fields []ds.Field
	for fieldRows.Next() {
		var field ds.Field

		err = fieldRows.Scan(&field.Type, &field.Name, &field.Value)
		if err != nil {
			return cipher, err
		}

		fields = append(fields, field)
	}

	cipher.Fields = fields

	attachmentRows, err := db.Query("SELECT id, filename, key, size, url FROM attachments WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	var attachments []ds.Attachment
	for attachmentRows.Next() {
		var attachment ds.Attachment
		err = attachmentRows.Scan(&attachment.Id, &attachment.FileName, &attachment.Key, &attachment.Size, &attachment.Url)
		if err != nil {
			return cipher, err
		}
		attachment.Object = "attachment"
		size, err := strconv.Atoi(attachment.Size)
		if err != nil {
			return cipher, err
		}
		attachment.SizeName = strconv.FormatInt(int64(size>>10), 10) + " KB"
		attachments = append(attachments, attachment)
	}
	cipher.Attachments = attachments

	cardRow, err := db.Query("SELECT cardholdername, brand, number, expmonth, expyear, code FROM cards WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	err = cardRow.Scan(&cipher.Card.CardHolderName, &cipher.Card.Brand, &cipher.Card.Number, &cipher.Card.ExpMonth, &cipher.Card.ExpYear, &cipher.Card.Code)
	if err != nil {
		return cipher, err
	}

	identityRow, err := db.Query("SELECT * FROM identities WHERE cipherId=$1", cipher.Id)
	if err != nil {
		return cipher, err
	}
	var foo, bar string
	err = identityRow.Scan(
		&foo,
		&bar,
		&cipher.Identity.Title,
		&cipher.Identity.FirstName,
		&cipher.Identity.MiddleName,
		&cipher.Identity.LastName,
		&cipher.Identity.Address1,
		&cipher.Identity.Address2,
		&cipher.Identity.Address3,
		&cipher.Identity.City,
		&cipher.Identity.State,
		&cipher.Identity.PostalCode,
		&cipher.Identity.Country,
		&cipher.Identity.Company,
		&cipher.Identity.Email,
		&cipher.Identity.Phone,
		&cipher.Identity.Ssn,
		&cipher.Identity.Username,
		&cipher.Identity.PassportNumber,
		&cipher.Identity.LicenseNumber)
	if err != nil {
		return cipher, err
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

func (db *DB) Init() error {
	if utils.PathExist(dbFileName) {
		err := os.Remove(dbFileName)
		if err != nil {
			return err
		}
	}

	for _, sql := range []string{identityTable, cardTable, accountTable, folderTable, cipherTable, loginTable, uriTable, fieldTable, attachmentTable} {
		if _, err := db.db.Exec(sql); err != nil {
			return errors.New(fmt.Sprintf("Sql error with %s\n%s", sql, err.Error()))
		}
	}
	return nil
}
