package server

import (
	"fmt"
	"time"
)

type Config struct {
	// api
	JwtSecret    string
	addr         string
	port         int
	readtimeout  time.Duration
	writetimeout time.Duration
	debug        bool
	// database
	dbHost   string
	dbPort   int
	dbDriver DriverType
	dbName   string
	dbUser   string
	dbPass   string
}

func NewConfig(opts ...ConfigOption) *Config {
	const (
		defAddr         = "localhost"
		defPort         = 8080
		defReadtimeout  = time.Second * 10
		defWritetimeout = time.Second * 10
		defDebug        = false
		defDbHost       = "localhost"
		defDbType       = SqliteMem
		defDbName       = "shiftr"
		defJtwSecret    = "changemeohgodplease"
	)

	c := &Config{
		addr:         defAddr,
		port:         defPort,
		readtimeout:  defReadtimeout,
		writetimeout: defWritetimeout,
		debug:        defDebug,
		dbHost:       defDbHost,
		dbDriver:     defDbType,
		dbName:       defDbName,
		JwtSecret:    defJtwSecret,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Config) serverURL() string {
	return fmt.Sprintf("%s:%d", c.addr, c.port)
}

const (
	SqliteMemoryUrl    = "file::memory:?cache=shared"
	SqliteUrlFormat    = "%s.db"
	MysqlUrlFormat     = "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	PostgresUrlFormat  = "host=%s port=%d user=%s dbname=%s sslmode=disable password=%s"
	SqlServerUrlFormat = "sqlserver://%s:%s@%s:%d?database=%s"
)

func (c *Config) databaseUrl() string {
	switch c.dbDriver {
	case SqliteMem:
		return SqliteMemoryUrl
	case Sqlite:
		return fmt.Sprintf(SqliteUrlFormat, c.dbName)
	case Postgres:
		return fmt.Sprintf(PostgresUrlFormat, c.dbHost, c.dbPort, c.dbUser, c.dbName, c.dbPass)
	case Mysql:
		return fmt.Sprintf(MysqlUrlFormat, c.dbUser, c.dbPass, c.dbHost, c.dbPort, c.dbName)
	case Sqlserver:
		return fmt.Sprintf(SqlServerUrlFormat, c.dbUser, c.dbPass, c.dbHost, c.dbPort, c.dbName)
	}

	return ""
}

type ConfigOption func(*Config)

func ListenPort(port int) ConfigOption {
	return func(c *Config) {
		c.port = port
	}
}

func ListenAddr(addr string) ConfigOption {
	return func(c *Config) {
		c.addr = addr
	}
}

type DriverType string

const (
	SqliteMem DriverType = "sqlitemem"
	Sqlite = "sqlite"
	Postgres = "postgres"
	Mysql = "mysql"
	Sqlserver = "sqlserver"
)

func (d DriverType) String() string {
	return string(d)
}

func GetDriverType(val string) DriverType {
	switch val {
	case "sqlitemem":
		return SqliteMem
	case "sqlite":
		return Sqlite
	case "postgres":
		return Postgres
	case "mysql":
		return Mysql
	case "sqlserver":
		return Sqlserver
	default:
		return SqliteMem
	}
}

func DatabaseDriver(db DriverType) ConfigOption {
	return func(c *Config) {
		c.dbDriver = db
	}
}

func DatabaseHost(host string) ConfigOption {
	return func(c *Config) {
		c.dbHost = host
	}
}

func DatabasePort(port int) ConfigOption {
	return func(c *Config) {
		c.dbPort = port
	}
}

func DatabaseName(name string) ConfigOption {
	return func(c *Config) {
		c.dbName = name
	}
}

func DatabaseUser(user string) ConfigOption {
	return func(c *Config) {
		c.dbUser = user
	}
}

func DatabasePass(pass string) ConfigOption {
	return func(c *Config) {
		c.dbPass = pass
	}
}

func WithReadTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.readtimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.writetimeout = timeout
	}
}

func DebugEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.debug = enabled
	}
}
