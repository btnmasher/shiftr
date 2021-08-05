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

// NewConfig returns a prepared Config struct with the given ConfigOption parameters modifying the state.
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

// ListenPort sets the port which the http server will accept connection. Default: 8080
func ListenPort(port int) ConfigOption {
	return func(c *Config) {
		c.port = port
	}
}

// ListenAddr sets the port which the http server will accept connection. Default: localhost
func ListenAddr(addr string) ConfigOption {
	return func(c *Config) {
		c.addr = addr
	}
}

// WithJWTSecret sets the JWT secret key to use for authentication. (CHANGE THE DEFAULT!) Default: changemeohgodplease
func WithJWTSecret(secret string) ConfigOption {
	return func(c *Config) {
		c.JwtSecret = secret
	}
}

type DriverType string

const (
	SqliteMem DriverType = "sqlitemem"
	Sqlite               = "sqlite"
	Postgres             = "postgres"
	Mysql                = "mysql"
	Sqlserver            = "sqlserver"
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

// DatabaseDriver sets the determined database you wish to use for persisting the data. Default: Sqlite Memory
func DatabaseDriver(db DriverType) ConfigOption {
	return func(c *Config) {
		c.dbDriver = db
	}
}

// DatabaseHost sets the host address for the database you wish to connect to. Default: localhost
func DatabaseHost(host string) ConfigOption {
	return func(c *Config) {
		c.dbHost = host
	}
}

// DatabasePort sets the port for the database you wish toc connect to. Default: none
func DatabasePort(port int) ConfigOption {
	return func(c *Config) {
		c.dbPort = port
	}
}

// DatabaseName sets the name of the database. Default: shiftr
func DatabaseName(name string) ConfigOption {
	return func(c *Config) {
		c.dbName = name
	}
}

// DatabaseUser sets the user to log into the database with. Default: none
func DatabaseUser(user string) ConfigOption {
	return func(c *Config) {
		c.dbUser = user
	}
}

//DatabasePass sets the password to log into the database with. Default: none
func DatabasePass(pass string) ConfigOption {
	return func(c *Config) {
		c.dbPass = pass
	}
}

// WithReadTimeout sets the http Read Timeout. Default: time.Second * 10
func WithReadTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.readtimeout = timeout
	}
}

// WithWriteTimeout sets the http Read Timeout. Default: time.Second * 10
func WithWriteTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.writetimeout = timeout
	}
}

// DebugEnabled sets whether or not to enable Debug logging (sensitive data will be written to stdout!). Default: false
func DebugEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.debug = enabled
	}
}
