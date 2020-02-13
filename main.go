package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/404cn/gowarden/api"
	"github.com/404cn/gowarden/database"
)

var gowarden struct {
	initDB              bool
	dir                 string
	port                string
	disableRegistration bool
}

func init() {
	flag.BoolVar(&gowarden.initDB, "initDB", false, "Initalizes the databases.")
	flag.StringVar(&gowarden.dir, "d", "", "Set the directory.")
	flag.StringVar(&gowarden.port, "p", "9527", "Set the Port.")
	flag.BoolVar(&gowarden.disableRegistration, "disableRegistration", false, "Disable registration.")
}

func main() {
	flag.Parse()

	database.StdDB.SetDir(gowarden.dir)

	database.StdDB.Open()
	defer database.StdDB.Close()

	// just for test
	// TODO delete
	gowarden.initDB = true

	if gowarden.initDB {
		log.Println("Try to initalize database ...")
		err := database.StdDB.Init()
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Database initalized.")
	}

	server := &http.Server{
		Addr: "127.0.0.1:9527",
	}

	if !gowarden.disableRegistration {
		http.HandleFunc("/api/accounts/register", api.HandleRegister)
	}

	server.ListenAndServe()
}
