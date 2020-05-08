package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/404cn/gowarden/api"
	"github.com/404cn/gowarden/sqlite"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

var gowarden struct {
	initDB              bool
	dir                 string
	port                string
	disableRegistration bool
	secretKey           string
	logLevel            string
	logPath             string
	disableFavicon      bool
	faviconProxyServer  string
	enableHttps         bool
	cert                string
	key                 string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.BoolVar(&gowarden.initDB, "initDB", false, "Initalizes the database.")
	flag.StringVar(&gowarden.dir, "d", "", "Set the directory.")
	flag.StringVar(&gowarden.port, "p", "9527", "Set the Port.")
	flag.BoolVar(&gowarden.disableRegistration, "disableRegistration", false, "Disable registration.")
	flag.StringVar(&gowarden.secretKey, "secertKey", "secret", "Use to encrypt jwt string.")
	flag.StringVar(&gowarden.logLevel, "loglevel", "Info", "Set log level.")
	flag.StringVar(&gowarden.logPath, "logpath", "", "Set log path.")
	flag.BoolVar(&gowarden.disableFavicon, "disableFavicon", false, "Disable favicon server.")
	// TODO change default to empty.
	flag.StringVar(&gowarden.faviconProxyServer, "faviconProxyServer", "http://127.0.0.1:7890", "Set favicon's proxy server.")
	flag.BoolVar(&gowarden.enableHttps, "enableHttps", false, "Set true to enable https.")
	flag.StringVar(&gowarden.cert, "certFile", "", "Path to cert.pem file")
	flag.StringVar(&gowarden.key, "keyFile", "", "Path to key.pem file.")
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func main() {
	flag.Parse()

	// TODO set log module
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer sugar.Sync()

	// TODO set log level and path

	db := sqlite.New()
	db.SetDir(gowarden.dir)
	err := db.Open()
	if err != nil {
		sugar.Fatal(err)
		return
	}
	defer db.Close()

	// just for test TODO delete
	// gowarden.initDB = true

	if gowarden.initDB {
		// TODO delete icon and attachment folders
		sugar.Info("Try to initialize sqlite ...")
		err := db.Init()
		if err != nil {
			logrus.Fatal(err)
			return
		}
		sugar.Info("Database initialized.")
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
		if !isDir("icons") {
			sugar.Info("Didn't find icon's cache folder, try to create...")
			err = os.Mkdir("icons", os.ModePerm)
			if err != nil {
				sugar.Error(err)
			}
			sugar.Info("Success to create icons folder.")
		}
		r.HandleFunc("/icons/{domain}/{icon}", handler.HandleFavicon).Methods(http.MethodGet)
	}

	if !isDir("attachments") {
		sugar.Info("Didn't find attachments's folder, try to create ...")
		err = os.Mkdir("attachments", os.ModePerm)
		if err != nil {
			sugar.Error(err)
		}
		sugar.Info("Success to create attachments folder.")
	}
	r.HandleFunc("/api/ciphers/{cipherId}/attachment", handler.AuthMiddleware(handler.HandleAddAttachment)).Methods(http.MethodPost)
	r.HandleFunc("/api/ciphers/{cipherId}/attachment/{attachmentId}", handler.AuthMiddleware(handler.HandleDeleteAttachment)).Methods(http.MethodDelete)
	// FIXME don't know api endpoint
	r.HandleFunc("/attachments/{cipherId}/{attachmentId}", handler.HandleGetAttachment).Methods(http.MethodGet)

	if gowarden.enableHttps {
		log.Fatal(http.ListenAndServeTLS("127.0.0.1"+gowarden.port, gowarden.cert, gowarden.key, r))
	} else {
		log.Fatal(http.ListenAndServe("127.0.0.1:"+gowarden.port, r))
	}
}
