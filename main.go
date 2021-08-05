package main

import (
	"github.com/btnmasher/shiftr/server"
	"log"
)

func main() {
	cfg := server.NewConfig(
		server.DatabaseDriver(server.Sqlite),
		server.DebugEnabled(true),
	)

	srv := server.New()

	err := srv.Initialize(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	srv.Run()
}
