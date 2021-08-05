package middleware

import (
	"errors"
	"github.com/btnmasher/shiftr/api/models"
	"github.com/btnmasher/shiftr/utils"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type claims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.StandardClaims
}

func Login(c echo.Context) error {
	name := c.QueryParam("user")
	pass := c.QueryParam("pass")
	db := c.Get("db").(*gorm.DB)

	if name == "" || pass == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "you must provide valid credentials")
	}

	user, err := models.FindUserByName(db, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.ErrUnauthorized
		}

		return err
	}

	err = utils.VerifyPassword(user.Password, pass)
	if err != nil {
		return echo.ErrUnauthorized
	}

	// Set custom claims
	claims := &claims{
		user.ID,
		user.Name,
		user.Role,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := c.Get("jwtsecret").(string)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func UserAccessible(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return echo.ErrUnauthorized
		}

		token := user.(*jwt.Token)

		cl := token.Claims.(jwt.MapClaims)

		if cl["role"] != "user" && cl["role"] != "admin" {
			return echo.ErrUnauthorized
		}

		c.Set("id", cl["id"])
		c.Set("role", cl["role"])

		return next(c)
	}
}

func AdminAccessible(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user")
		if user == nil {
			return echo.ErrUnauthorized
		}

		token := user.(*jwt.Token)

		cl := token.Claims.(jwt.MapClaims)

		if cl["role"] != "admin" {
			return echo.ErrUnauthorized
		}

		c.Set("id", cl["id"])
		c.Set("role", cl["role"])

		return next(c)
	}
}

func TestAccessible(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		c.Set("role", "admin")

		return next(c)
	}
}
