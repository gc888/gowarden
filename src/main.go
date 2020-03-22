package main

import (
	"flag"
	"log"
	"net/http"

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
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.BoolVar(&gowarden.initDB, "initDB", false, "Initalizes the database.")
	flag.StringVar(&gowarden.dir, "d", "", "Set the directory.")
	flag.StringVar(&gowarden.port, "p", "9527", "Set the Port.")
	flag.BoolVar(&gowarden.disableRegistration, "disableRegistration", false, "Disable registration.")
	flag.StringVar(&gowarden.secretKey, "key", "secret", "Use to encrypt jwt string.")
	flag.StringVar(&gowarden.logLevel, "loglevel", "Info", "Set log level.")
	flag.StringVar(&gowarden.logPath, "logpath", "", "Set log path.")
}

func main() {
	flag.Parse()

	// TODO set log module
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer sugar.Sync()

	// TODO just for test
	gowarden.initDB = true

	// TODO set log level and path

	sqlite.StdDB.SetDir(gowarden.dir)
	err := sqlite.StdDB.Open()
	if err != nil {
		sugar.Fatal(err)
		return
	}
	defer sqlite.StdDB.Close()

	if gowarden.initDB {
		sugar.Info("Try to initalize sqlite ...")
		err := sqlite.StdDB.Init()
		if err != nil {
			logrus.Fatal(err)
			return
		}
		sugar.Info("Database initalized.")
	}

	r := mux.NewRouter()
	handler := api.New(sqlite.StdDB, gowarden.secretKey, sugar)

	if !gowarden.disableRegistration {
		r.HandleFunc("/api/accounts/register", handler.HandleRegister)
	}

	r.HandleFunc("/api/accounts/prelogin", handler.HandlePrelogin)
	r.HandleFunc("/identity/connect/token", handler.HandleLogin)

	// Must login can access these api.
	r.HandleFunc("/api/accounts/keys", handler.AuthMiddleware(handler.HandleAccountKeys))
	// TODO
	r.HandleFunc("/api/sync", handler.AuthMiddleware(handler.HandleSync))
	r.HandleFunc("/notifications/hub/negotiate", handler.AuthMiddleware(handler.HandleNegotiate))
	r.HandleFunc("/api/ciphers", handler.AuthMiddleware(handler.HandleCiphers))

	r.HandleFunc("/api/folders", handler.AuthMiddleware(handler.HandleFolder)).Methods(http.MethodPost)
	r.HandleFunc("/api/folders/{folderUUID}", handler.AuthMiddleware(handler.HandleFolderRename)).Methods(http.MethodPut)
	r.HandleFunc("/api/folders/{folderUUID}", handler.AuthMiddleware(handler.HandleFolderDelete)).Methods(http.MethodDelete)

	log.Fatal(http.ListenAndServe("127.0.0.1:"+gowarden.port, r))
}
