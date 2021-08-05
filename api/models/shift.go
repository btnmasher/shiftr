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
	shifts, err := ListShifts(db, s.UserID, -1, s.Start, s.End)

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

// ListShifts attempts to return rows from the Shifts table with the specified limits and filters ordered by start time
// If uid is not an empty string, results will be filtered to those rows belonging to any matching User.ID
// If limit specified is less than or equal to 0, result will not be limited
// The results can be filtered by the limitStart or limitEnd time, leave either fields nil for no filtering
func ListShifts(db *gorm.DB, uid string, limit int, filterStart, filterEnd time.Time) ([]*Shift, error) {
	var shifts []*Shift

	tx := db.Model(&Shift{}).Order("start")

	if uid != "" {
		tx.Where("user_id = ?", uid)
	}

	if !filterStart.IsZero() {
		tx.Where("start >= ?", filterStart)
	}

	if !filterEnd.IsZero() {
		tx.Where("end <= ?", filterEnd)
	}

	if limit < 1 {
		limit = -1
	}

	tx.Limit(limit).Find(&shifts)

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
