package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/404cn/gowarden/api"
	"github.com/404cn/gowarden/sqlite"
)

var gowarden struct {
	initDB              bool
	dir                 string
	port                string
	disableRegistration bool
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.BoolVar(&gowarden.initDB, "initDB", false, "Initalizes the database.")
	flag.StringVar(&gowarden.dir, "d", "", "Set the directory.")
	flag.StringVar(&gowarden.port, "p", "9527", "Set the Port.")
	flag.BoolVar(&gowarden.disableRegistration, "disableRegistration", false, "Disable registration.")
}

func main() {
	flag.Parse()

	gowarden.initDB = true

	sqlite.StdDB.SetDir(gowarden.dir)

	err := sqlite.StdDB.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer sqlite.StdDB.Close()

	if gowarden.initDB {
		log.Println("Try to initalize sqlite ...")
		err := sqlite.StdDB.Init()
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Database initalized.")
	}

	server := &http.Server{
		Addr: "127.0.0.1:9527",
		// Addr: "127.0.0.1:" + gowarden.port,
	}

	handler := api.StdApiHandler

	if !gowarden.disableRegistration {
		http.HandleFunc("/api/accounts/register", handler.HandleRegister)
	}

	http.HandleFunc("/api/accounts/prelogin", handler.HandlePrelogin)
	http.HandleFunc("/identity/connect/token", handler.HandleLogin)

	// Must login can access these api.
	http.HandleFunc("/api/accounts/keys", handler.AuthMiddleware(handler.HandleAccountKeys))
	http.HandleFunc("/api/sync", handler.AuthMiddleware(handler.HandleSync))
	http.HandleFunc("/notifications/hub/negotiate", handler.AuthMiddleware(handler.HandleNegotiate))

	http.HandleFunc("/api/folders", handler.AuthMiddleware(handler.HandlerFolders))
	http.HandleFunc("/api/ciphers", handler.AuthMiddleware(handler.HandleCiphers))

	log.Fatal(server.ListenAndServe())

}
