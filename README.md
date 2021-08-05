# shiftr

A coding challenge to demonstrate my knowledge with Go, REST, Patterns, and Databases

## Running:

`go get` and build! It defaults to using Sqlite memory database, some demo data has been added

There is a postman collection file added for testing the endpoints.

##Build Dependencies

Requires GCC to build the sqlite dependency of GORM

If running on Windows, install [tdm-gcc](https://jmeubank.github.io/tdm-gcc/) or equivalent so that there's a GCC binary in your %PATH%

##Customizaton

You can customize/configure the application with the provided configuration functions passed to `server.NewCOnfig()`

```Go
package main

import (
	"github.com/btnmasher/shiftr/server"
	"log"
)

func main() {
	cfg := server.NewConfig(
                server.ListenAddr("localhost"),
                server.ListenPort(8080),
                server.WithJWTSecret("a strong secret here!"),
                server.DatabaseDriver(server.Postgres),
                server.DatabaseHost("localhost"),
                server.DatabasePort(5432),
                server.DatabaseUser("postgres_user"),
                server.DatabasePass("postgres_password"),
                server.DatabaseName("shiftr"),
                server.WithReadTimeout(time.Second * 5),
                server.WithWriteTimeout(time.Second * 5),
                server.DebugEnabled(true),
	)

	srv := server.New()

	err := srv.Initialize(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	srv.Run()
}
```
