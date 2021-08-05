package main

import (
	"github.com/btnmasher/shiftr/api/models"
	"github.com/btnmasher/shiftr/server"
	"log"
	"time"
)

func main() {
	cfg := server.NewConfig(
		server.DatabaseDriver(server.SqliteMem),
		server.DebugEnabled(true),
	)

	srv := server.New()

	err := srv.Initialize(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	setupDemoData(srv)

	srv.Run()
}

func setupDemoData(srv *server.Server) {
	db := srv.DB

	admin := &models.User{
		Name:     "adminuser",
		Password: "adminpass",
		Role:     "admin",
	}

	admin.Create(db)

	user := &models.User{
		Name:     "testuser",
		Password: "testpass",
		Role:     "user",
	}

	user.Create(db)

	shift := &models.Shift{
		Start:  time.Now(),
		End:    time.Now().Add(time.Hour * 8),
		UserID: user.ID,
	}

	shift.Create(db)

}
