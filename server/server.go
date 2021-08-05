package server

import (
	"errors"
	"fmt"
	"github.com/btnmasher/shiftr/api/handlers"
	"github.com/btnmasher/shiftr/api/middleware"
	"github.com/btnmasher/shiftr/api/models"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

type Server struct {
	DB     *gorm.DB
	API    *echo.Echo
	Config *Config
}

func New() *Server {
	return &Server{}
}

// Initialize starts the Server, connecting to the database specified in the configuration
// and setting up the defined API routes.
func (s *Server) Initialize(config *Config) error {
	s.Config = config

	cfg := &gorm.Config{}

	if config.debug {
		cfg.Logger = logger.Default.LogMode(logger.Info)
		fmt.Printf("Configuration Initializing:\n%+v\n", *config)
	}

	var err error
	switch config.dbDriver {
	case SqliteMem:
		fallthrough
	case Sqlite:
		s.DB, err = gorm.Open(sqlite.Open(config.databaseUrl()), cfg)
		break
	case Postgres:
		s.DB, err = gorm.Open(postgres.Open(config.databaseUrl()), cfg)
		break
	case Mysql:
		s.DB, err = gorm.Open(mysql.Open(config.databaseUrl()), cfg)
		break
	case Sqlserver:
		s.DB, err = gorm.Open(sqlserver.Open(config.databaseUrl()), cfg)
		break
	default:
		return errors.New("unknown/unsupported database driver type specified")
	}

	if err != nil {
		return fmt.Errorf("could not connect to %s database: %s", config.dbDriver, err)
	}

	log.Printf("connected to the %s database successfully", config.dbDriver)

	err = s.DB.AutoMigrate(&models.User{}, &models.Shift{}) //database migration
	if err != nil {
		return fmt.Errorf("could not automigrate models: %s", err)
	}

	log.Printf("migrated %s database models successfully", config.dbDriver)

	s.API = echo.New()
	s.API.HideBanner = true
	s.API.Debug = config.debug
	s.API.Server.ReadTimeout = config.readtimeout
	s.API.Server.WriteTimeout = config.writetimeout

	s.API.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("jwtsecret", config.JwtSecret)
			c.Set("db", s.DB)
			return next(c)
		}
	})

	s.API.Use(echomw.Logger())

	s.initRoutes()

	return nil
}

func (s *Server) initRoutes() {
	s.API.POST("/login", middleware.Login)

	// Wrap the /api/v1 route in JWT auth
	g := s.API.Group("/api/v1")
	g.Use(echomw.JWT([]byte(s.Config.JwtSecret)))

	// User-role accessible endpoints
	g.GET("/shifts", handlers.ListShifts(), middleware.UserAccessible)
	g.GET("/shifts/:id", handlers.GetShift(), middleware.UserAccessible)
	g.POST("/shifts", handlers.CreateShift(), middleware.UserAccessible)
	g.PUT("/shifts/:id", handlers.UpdateShift(), middleware.UserAccessible)
	g.DELETE("/shifts/:id", handlers.DeleteShift(), middleware.UserAccessible)
	g.GET("/users/:id", handlers.GetUserByID(), middleware.UserAccessible)
	g.PUT("/users/:id", handlers.UpdateUser(), middleware.UserAccessible)

	// Admin-role accessible endpoints
	g.GET("/users", handlers.ListUsers(), middleware.AdminAccessible)
	g.POST("/users", handlers.CreateUser(), middleware.AdminAccessible)
	g.DELETE("/users/:id", handlers.DeleteUser(), middleware.AdminAccessible)
}

func (s *Server) Run() {
	s.API.Logger.Fatal(s.API.Start(s.Config.serverURL()))
}
