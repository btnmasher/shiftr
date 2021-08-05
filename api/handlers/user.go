package handlers

import (
	"errors"
	"github.com/btnmasher/shiftr/api/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateUser() func(echo.Context) error {
	return func(c echo.Context) error {

		// Collect the submitted data from the user
		data := &models.User{}
		err := c.Bind(data)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"invalid binding")
		}

		// Prepare a new object to write to the database
		user := models.User{
			Name:     data.Name,
			Password: data.Password,
			Role:     data.Role,
		}

		// Ensure we have all necessary fields to create the object
		err = user.Validate()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Collect the database reference from context
		db := c.Get("db").(*gorm.DB)

		// Ensure there are no other users that already exist with the specified name
		_, err = models.FindUserByName(db, data.Name)
		if err == nil {
			return echo.NewHTTPError(http.StatusConflict, "user already exists")
		}

		// Return any other errors that have occurred that are not "record not found" errors.
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Attempt to write the new object to the database
		err = user.Create(db)
		if err != nil {
			return err
		}

		user.Password = ""

		return c.JSON(http.StatusCreated, user)
	}
}

func UpdateUser() func(echo.Context) error {
	return func(c echo.Context) error {

		// Collect the submitted data from the user
		data := &models.User{}
		err := c.Bind(data)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"invalid object")
		}

		// Collect context values
		role := c.Get("role").(string)
		uid := c.Get("id").(string)
		db := c.Get("db").(*gorm.DB)

		// Prepare a new object to write to the database
		change := models.User{
			ID:       uid,
			Name:     data.Name,
			Password: data.Password,
			Role:     data.Role,
		}

		// Ensure we have all necessary fields to update the object
		err = change.Validate()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Attempt to fetch the existing user object
		user, err := models.FindUserByID(db, uid)
		if err != nil {
			return echo.ErrNotFound
		}

		// Constrain the user from changing another user's object or their own role if not admin
		if role == "user" {
			if user.ID != uid {
				return echo.ErrUnauthorized
			}

			if change.Role != user.Role {
				return echo.ErrUnauthorized
			}
		}

		// Ensure that no other user exists with a matching name to the new changes
		if data.Name != user.Name {
			check, err := models.FindUserByName(db, data.Name)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// No match was found, change the name in the temporary object
					change.Name = data.Name
				} else {
					return nil
				}
			}

			if check.Name != "" {
				// Match found, forbid the request
				return echo.ErrForbidden
			}
		}

		// Attempt to write the change to the database
		err = change.Update(db)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		change.Password = ""

		return c.JSON(http.StatusOK, change)
	}
}

func ListUsers() func(echo.Context) error {
	return func(c echo.Context) error {
		//Safely ignoring error as an invalid limit parameter would return a zero, which is no limit for ListUsers
		limit, _ := strconv.Atoi(c.QueryParam("limit"))

		// Collect database reference from context
		db := c.Get("db").(*gorm.DB)

		// Attempt to list the users rom the database
		users, err := models.ListUsers(db, limit)
		if err != nil {
			return err
		}

		// Clear sensitive information from the returned objects
		for i := range users {
			users[i].Password = ""
		}

		return c.JSON(http.StatusOK, users)
	}
}

func GetUserByID() func(ctx echo.Context) error {
	return func(c echo.Context) error {

		// Collect parameters and conext values
		id := c.Param("id")
		db := c.Get("db").(*gorm.DB)
		role := c.Get("role").(string)
		uid := c.Get("id").(string)

		// Attempt to find the user in the database with the specified ID
		user, err := models.FindUserByID(db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		// Constrain the user from fetching another user's object if not admin
		if role == "user" {
			if uid != user.ID {
				return echo.ErrUnauthorized
			}
		}

		// Clear sensitive information from the returned object
		user.Password = ""

		return c.JSON(http.StatusOK, user)
	}
}

func DeleteUser() func(ctx echo.Context) error {
	return func(c echo.Context) error {

		// Collect parameters and context values
		uid := c.Param("id")
		db := c.Get("db").(*gorm.DB)

		// Attempt to find the user in the database wit the specified ID
		user, err := models.FindUserByID(db, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		// Attempt to delete the object from the database
		err = user.Delete(db)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
