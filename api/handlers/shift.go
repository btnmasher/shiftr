package handlers

import (
	"errors"
	"github.com/btnmasher/shiftr/api/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func CreateShift() func(echo.Context) error {
	return func(c echo.Context) error {

		// Collect the submitted data from the user
		data := &models.Shift{}
		err := c.Bind(data)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid object")
		}

		// Prepare a new object to write to the database
		shift := models.Shift{
			UserID: data.UserID,
			Start:  data.Start,
			End:    data.End,
		}

		// Ensure we have all necessary fields to create the object
		err = shift.Validate()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Collect context values
		role := c.Get("role").(string)
		uid := c.Get("id").(string)

		// Constrain the user from creating a shift object for another user if not admin
		if role == "user" {
			if uid != shift.UserID {
				return echo.ErrUnauthorized
			}
		}

		// Collect the database reference from context
		db := c.Get("db").(*gorm.DB)

		// Attempt to write the new object to the database
		err = shift.Create(db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, shift)
	}
}

func UpdateShift() func(echo.Context) error {
	return func(c echo.Context) error {

		// Collect the submitted data from the user
		data := &models.Shift{}
		err := c.Bind(data)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid object")
		}

		// Collect parameters and context values
		sid := c.Param("id")
		role := c.Get("role").(string)
		uid := c.Get("id").(string)
		db := c.Get("db").(*gorm.DB)

		// Check if the shift already exists
		shift, err := models.FindShiftByID(db, sid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		// Constrain the user from changing the UserID of the Shift object if not admin
		if role == "user" {
			if shift.UserID != uid {
				return echo.ErrUnauthorized
			}

			if data.UserID != "" && data.UserID != uid {
				return echo.ErrUnauthorized
			}
		}

		// Prepare a new object to write to the database
		change := models.Shift{
			ID:     sid,
			UserID: data.UserID,
			Start:  data.Start,
			End:    data.End,
		}

		// Ensure there are no zero values before writing
		if data.UserID == "" {
			change.UserID = shift.UserID
		}

		if data.Start.IsZero() {
			change.Start = shift.End
		}

		if data.End.IsZero() {
			change.End = shift.End
		}

		// Attempt to write the new object to the database
		err = change.Update(db)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, change)
	}
}

func ListShifts() func(echo.Context) error {
	return func(c echo.Context) error {

		// A temporary struct to hold our user submitted data for binding
		var params struct {
			UserID string    `query:"user_id"`
			Start  time.Time `query:"filter_start"` // RFC33339
			End    time.Time `query:"filter_end"`   // RFC33339
			Limit  int       `query:"limit"`
		}

		// Collect the submitted data from the user
		err := c.Bind(&params)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"invalid parameters")
		}

		// Collect context values
		role := c.Get("role").(string)
		uid := c.Get("id").(string)

		// Constrain the user from listing shifts from another user if not admin
		if role == "user" {
			if uid != params.UserID {
				if params.UserID == "" {
					// Ensure the user only receives relevant results for their UserID
					params.UserID = uid
				} else {
					return echo.ErrUnauthorized
				}
			}
		}

		// Ensure that the timestamp received isn't malformed
		if !params.Start.IsZero() && !params.End.IsZero() {
			if params.Start.After(params.End) {
				return echo.NewHTTPError(http.StatusBadRequest,
					"filter span start time must precede span end time")
			}
		}

		// Collect database reference from context
		db := c.Get("db").(*gorm.DB)

		// Attempt to write the changes to the database
		shifts, err := models.ListShifts(db,
			models.FilterUserID(params.UserID),
			models.FilterStart(params.Start),
			models.FilterEnd(params.End),
			models.WithLimit(params.Limit),
		)
		//shifts, err := models.ListShifts(db, params.UserID, params.Limit, params.Start, params.End)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, shifts)
	}
}

func GetShift() func(ctx echo.Context) error {
	return func(c echo.Context) error {

		// Collect parameters and context values
		sid := c.Param("id")
		db := c.Get("db").(*gorm.DB)
		role := c.Get("role").(string)
		uid := c.Get("id").(string)

		// Attempt to find the shift in the database
		shift, err := models.FindShiftByID(db, sid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		// Constrain the user from fetching shifts that do not match their UserID if not admin
		if role == "user" {
			if uid != shift.UserID {
				return echo.ErrUnauthorized
			}
		}

		return c.JSON(http.StatusOK, shift)
	}
}

func DeleteShift() func(ctx echo.Context) error {
	return func(c echo.Context) error {

		// Collect parameters and context values
		sid := c.Param("id")
		db := c.Get("db").(*gorm.DB)
		role := c.Get("role").(string)
		uid := c.Get("id").(string)

		// Attempt to find the shift in the database
		shift, err := models.FindShiftByID(db, sid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.ErrNotFound
			}

			return err
		}

		// Constrain the user from deleting shifts that do not match their UserID if not admin
		if role == "user" {
			if uid != shift.UserID {
				return echo.ErrUnauthorized
			}
		}

		// Attempt to delete the object from the database
		err = shift.Delete(db)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	}
}
