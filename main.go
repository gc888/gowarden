package main

import (
	"encoding/csv"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/404cn/gowarden/ds"
	"github.com/404cn/gowarden/logger"
	"github.com/404cn/gowarden/utils"

	"github.com/404cn/gowarden/api"
	"github.com/404cn/gowarden/sqlite"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var gowarden struct {
	initDB              bool
	dir                 string
	port                string
	disableRegistration bool
	secretKey           string
	logLevel            int
	disableFavicon      bool
	faviconProxyServer  string
	enableHttps         bool
	cert                string
	key                 string
	csvFile             string
	username            string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.BoolVar(&gowarden.initDB, "initDB", false, "Initalizes the database.")
	flag.StringVar(&gowarden.dir, "d", "", "Set the directory.")
	flag.StringVar(&gowarden.port, "p", "9527", "Set the Port.")
	flag.BoolVar(&gowarden.disableRegistration, "disableRegistration", false, "Disable registration.")
	flag.StringVar(&gowarden.secretKey, "secertKey", "secret", "Use to encrypt jwt string.")
	// TODO set level to info
	flag.IntVar(&gowarden.logLevel, "loglevel", -1, "Set log level, default is info.")
	flag.BoolVar(&gowarden.disableFavicon, "disableFavicon", false, "Disable favicon server.")
	// TODO change default to empty.
	flag.StringVar(&gowarden.faviconProxyServer, "faviconProxyServer", "http://127.0.0.1:7890", "Set favicon's proxy server.")
	flag.BoolVar(&gowarden.enableHttps, "enableHttps", false, "Set true to enable https.")
	flag.StringVar(&gowarden.cert, "certFile", "", "Path to cert.pem file")
	flag.StringVar(&gowarden.key, "keyFile", "", "Path to key.pem file.")
	flag.StringVar(&gowarden.csvFile, "csvFile", "", "Path to csv file.")
	// TODO change to default value
	flag.StringVar(&gowarden.username, "username or email", "nobody@example.com", "Only use with --csvFile to decide import data from csv to which account")
}

func main() {
	flag.Parse()
	var err error

	sugar, err := logger.New(gowarden.logLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer sugar.Sync()

	db := sqlite.StdDB
	db.SetDir(gowarden.dir)
	err = db.Open()
	if err != nil {
		sugar.Fatal(err)
		return
	}
	defer db.Close()

	// just for test TODO delete
	// gowarden.initDB = true

	if gowarden.initDB {
		sugar.Info("Try to initialize database ...")
		err := db.Init()
		if err != nil {
			logrus.Fatal(err)
			return
		}
		sugar.Info("Database initialized.")
	}

	// TODO test
	if gowarden.csvFile != "" {
		csvs, err := importFromCSV(gowarden.csvFile, gowarden.username)
		if err != nil {
			log.Fatal(err)
		}

		err = db.SaveCSV(csvs)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := mux.NewRouter()
	handler := api.New(db, gowarden.secretKey, sugar, gowarden.faviconProxyServer)

	if !gowarden.disableRegistration {
		r.HandleFunc("/api/accounts/register", handler.HandleRegister)
	}

	r.HandleFunc("/api/accounts/prelogin", handler.HandlePrelogin)
	r.HandleFunc("/identity/connect/token", handler.HandleLogin)

	// Must login can access these api.
	r.HandleFunc("/api/accounts/keys", handler.AuthMiddleware(handler.HandleAccountKeys))
	r.HandleFunc("/api/sync", handler.AuthMiddleware(handler.HandleSync)).Methods(http.MethodGet)
	r.HandleFunc("/notifications/hub/negotiate", handler.AuthMiddleware(handler.HandleNegotiate))
	r.HandleFunc("/api/ciphers", handler.AuthMiddleware(handler.HandleCiphers)).Methods(http.MethodPost)
	r.HandleFunc("/api/ciphers/{cipherId}", handler.AuthMiddleware(handler.HandleUpdateCiphers)).Methods(http.MethodPut)
	r.HandleFunc("/api/ciphers/{cipherId}", handler.AuthMiddleware(handler.HandleDeleteCiphers)).Methods(http.MethodDelete)

	r.HandleFunc("/api/folders", handler.AuthMiddleware(handler.HandleFolder)).Methods(http.MethodPost)
	r.HandleFunc("/api/folders/{folderUUID}", handler.AuthMiddleware(handler.HandleFolderRename)).Methods(http.MethodPut)
	r.HandleFunc("/api/folders/{folderUUID}", handler.AuthMiddleware(handler.HandleFolderDelete)).Methods(http.MethodDelete)

	if !gowarden.disableFavicon {
		if !utils.IsDir("icons") {
			sugar.Info("Didn't find icon's cache folder, try to create...")
			err = os.Mkdir("icons", os.ModePerm)
			if err != nil {
				sugar.Error(err)
			}
			sugar.Info("Success to create icons folder.")
		}
		r.HandleFunc("/icons/{domain}/{icon}", handler.HandleFavicon).Methods(http.MethodGet)
	}

	if !utils.IsDir("attachments") {
		sugar.Info("Didn't find attachments's folder, try to create ...")
		err = os.Mkdir("attachments", os.ModePerm)
		if err != nil {
			sugar.Error(err)
		}
		sugar.Info("Success to create attachments folder.")
	}
	r.HandleFunc("/api/ciphers/{cipherId}/attachment", handler.AuthMiddleware(handler.HandleAddAttachment)).Methods(http.MethodPost)
	r.HandleFunc("/api/ciphers/{cipherId}/attachment/{attachmentId}", handler.AuthMiddleware(handler.HandleDeleteAttachment)).Methods(http.MethodDelete)
	r.HandleFunc("/attachments/{cipherId}/{attachmentId}", handler.HandleGetAttachment).Methods(http.MethodGet)

	if gowarden.enableHttps {
		log.Fatal(http.ListenAndServeTLS("127.0.0.1"+gowarden.port, gowarden.cert, gowarden.key, r))
	} else {
		log.Fatal(http.ListenAndServe("127.0.0.1:"+gowarden.port, r))
	}
}

func importFromCSV(file, email string) ([]ds.CSV, error) {
	var csvs []ds.CSV

	account, err := sqlite.StdDB.GetAccount(email)
	if err != nil {
		return csvs, err
	}

	fp, err := os.Open(file)
	if err != nil {
		return csvs, err
	}

	r := csv.NewReader(fp)
	records, err := r.ReadAll()
	if err != nil {
		return csvs, err
	}

	for foo, bar := range records {
		if foo == 0 {
			continue
		}
		var csv ds.CSV

		csv.Folder.Name = bar[0]
		csv.Folder.RevisionDate = time.Now()
		if bar[1] == "0" {
			csv.Favorite = false
		} else {
			csv.Favorite = true
		}
		csv.CipherType = bar[2]
		csv.Name = bar[3]
		csv.Notes = bar[4]
		fields := strings.Split(bar[5], "\n")
		for _, v := range fields {
			foo := strings.Split(v, ":")
			// FIXME if fields type is bool, then foo[1] maybe cause array out of index
			csv.Filds = append(csv.Filds, ds.Field{Name: foo[0], Value: foo[1]})
		}

		uris := strings.Split(bar[6], "\n")
		for _, v := range uris {
			csv.Login.Uris = append(csv.Login.Uris, ds.Uri{Uri: v})
		}
		csv.Login.Uri = csv.Login.Uris[0].Uri
		csv.Login.Username = &bar[7]
		csv.Login.Password = bar[8]
		csv.Login.Totp = bar[9]

		csvs = append(csvs, csv)
	}

	return csvs, nil
}
