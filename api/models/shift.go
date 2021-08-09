package models

import (
	"errors"
	"fmt"
	"github.com/jkomyno/nanoid"
	"gorm.io/gorm"
	"time"
)

// Shift struct represents a timespan of a work shift object with a Unique ID, Start and End times,
// and a UserID which the shift belongs to.
type Shift struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Start     time.Time `gorm:"not null" json:"start"`
	End       time.Time `gorm:"not null" json:"end"`
	UserID    string    `gorm:"not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks to ensure all fields of the object are present and valid
func (s *Shift) Validate() error {
	if s.Start.IsZero() {
		return errors.New("start time required")
	}

	if s.End.IsZero() {
		return errors.New("end time required")
	}

	if s.Start.After(s.End) {
		return errors.New("shift start time must precede shift end time")
	}

	return nil
}

// BeforeCreate hooks GORM and prepares a new object for creation
func (s *Shift) BeforeCreate(_ *gorm.DB) error {
	id, err := nanoid.Nanoid(10)
	if err != nil {
		return fmt.Errorf("unable to generate ShiftID: %s", err)
	}

	s.ID = id

	return nil
}

// BeforeSave hooks GORM to run necessary checks before saving the object
func (s *Shift) BeforeSave(db *gorm.DB) error {

	// Fetch all shifts that fall within the new shift's time span
	shifts, err := ListShifts(db,
		FilterUserID(s.UserID),
		FilterStart(s.Start),
		FilterEnd(s.End),
	)

	if err != nil {
		return err
	}

	overlap := false

	// Find overlapping shifts
	for _, shift := range shifts {
		if shift.End.After(s.Start) && shift.Start.Before(s.End) {
			if shift.ID != s.ID {
				overlap = true
			}
		}
	}

	if overlap {
		return errors.New("shift timespan cannot intersect other shifts for the same user")
	}

	return nil
}

// Create attempts to create the Shift object in the database
func (s *Shift) Create(db *gorm.DB) error {
	err := db.Create(s).Error
	if err != nil {
		return err
	}

	return nil
}

// Update will attempt to update the current Shift object in the database
func (s *Shift) Update(db *gorm.DB) error {

	// Update only the specific columns
	tx := db.Model(s).Where("id = ?", s.ID).Updates(
		map[string]interface{}{
			"start":   s.Start,
			"end":     s.End,
			"user_id": s.UserID,
		},
	).Take(s) // Update the current reference

	err := tx.Error
	if err != nil {
		return err
	}

	if tx.RowsAffected < 1 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete will attempt to delete the Shift object from the database
func (s *Shift) Delete(db *gorm.DB) error {
	tx := db.Delete(s)

	err := tx.Error
	if err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return errors.New("shift not found")
	}

	return nil
}

type ShiftFilterOption func(*gorm.DB)

// FilterUserID is used with ListShifts to filter the query to return results matching the specific User.ID
// If uid is not an empty string, results will be filtered by that User.ID
func FilterUserID(uid string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if uid != "" {
			db.Where("user_id = ?", uid)
		}
	}
}

// WithLimit is used with ListShifts to limit the number of results returned by the query.
// If limit specified is less than or equal to 0, result will not be limited
func WithLimit(limit int) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if limit < 1 {
			limit = -1
		}
		db.Limit(limit)
	}
}

// FilterStart is used with ListShifts to filter Shift results that have start times that fall on or before the
// specified filtered time.
// If start is specified as a time.Time zero value, it is ignored.
func FilterStart(start time.Time) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if !start.IsZero() {
			db.Where("start >= ?", start)
		}
	}
}

// FilterEnd is used with ListShifts to filter Shift results that have end times that fall on or after the
// specified filtered time.
// If end is specified as a time.Time zero value, it is ignored.
func FilterEnd(end time.Time) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if !end.IsZero() {
			db.Where("end <= ?", end)
		}
	}
}

// ListShifts attempts to return rows from the Shifts table with the specified limits and filters ordered by start time
// Provide ShiftFilterOption parameters to modify the query with additional filters.
func ListShifts(db *gorm.DB, opts ...ShiftFilterOption) ([]*Shift, error) {
	var shifts []*Shift

	tx := db.Model(&Shift{}).Order("start")

	for _, opt := range opts {
		opt(tx)
	}

	tx.Find(&shifts)

	err := tx.Error
	if err != nil {
		return []*Shift{}, err
	}

	return shifts, nil
}

// FindShiftByID attempts to return a row from the Shifts table with the matching ID
func FindShiftByID(db *gorm.DB, sid string) (*Shift, error) {

	shift := &Shift{}
	err := db.First(&shift, "id = ?", sid).Error
	if err != nil {
		return &Shift{}, err
	}

	return shift, nil
}
