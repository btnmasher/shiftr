package models

import (
	"errors"
	"fmt"
	"github.com/btnmasher/shiftr/utils"
	"github.com/jkomyno/nanoid"
	"gorm.io/gorm"
	"html"
	"strings"
	"time"
)

// User struct represents a user with a unique ID, Name, Password, and Role
type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:30;not null;unique'" json:"name"`        //login name
	Password  string    `gorm:"size:100;not null" json:"password,omitempty"` //bcrypt hash
	Role      string    `gorm:"size:10;not null" json:"role"`                //user role: user, admin
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks to ensure all fields of the object are present and valid
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name required")
	}

	if u.Password == "" {
		return errors.New("password required")
	}

	if u.Role == "" {
		return errors.New("role required")
	}

	if u.Role != "user" && u.Role != "admin" {
		return errors.New("invalid role")
	}

	return nil
}

// Prepare prepares a new object for update by escaping the name field and
// hashing the password field before it is written
func (u *User) Prepare() error {
	u.Name = html.EscapeString(strings.TrimSpace(u.Name))

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)

	return nil
}

// Create attempts to create the User object in the database
func (u *User) Create(db *gorm.DB) error {
	id, err := nanoid.Nanoid(8)
	if err != nil {
		return fmt.Errorf("unable to generate UserID: %s", err)
	}

	u.ID = id

	err = u.Prepare()
	if err != nil {
		return err
	}

	err = db.Create(u).Error
	if err != nil {
		return err
	}

	return nil
}

// Update will attempt to update the current User object in the database
func (u *User) Update(db *gorm.DB) error {
	err := u.Prepare()
	if err != nil {
		return err
	}

	// Update only the specific columns
	tx := db.Model(u).Where("id = ?", u.ID).Updates(
		map[string]interface{}{
			"name":     u.Name,
			"password": u.Password,
			"role":     u.Role,
		},
	).Take(u) // Update the current reference

	err = tx.Error
	if err != nil {
		return err
	}

	return nil
}

// Delete will attempt to delete the User object from the database
func (u *User) Delete(db *gorm.DB) error {
	tx := db.Delete(u)

	err := tx.Error
	if err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// AfterDelete hooks GORM to remove the associated Shift rows for ths user
// when it is deleted
func (u *User) AfterDelete(db *gorm.DB) error {
	return db.Model(&Shift{}).Where("user_id = ?", u.ID).Delete(&Shift{}).Error
}

// ListUsers attempts to return rows from the Users table with the specified limit
// If limit specified is less than or equal to 0, result will not be limited.
func ListUsers(db *gorm.DB, limit int) ([]*User, error) {
	var users []*User

	if limit < 1 {
		limit = -1
	}

	err := db.Model(&User{}).Limit(limit).Find(&users).Error
	if err != nil {
		return []*User{}, err
	}

	return users, nil
}

// FindUserByID attempts to return a row from the Users table with the matching User.ID
func FindUserByID(db *gorm.DB, uid string) (*User, error) {
	user := &User{}
	err := db.First(&user, "id = ?", uid).Error
	if err != nil {
		return &User{}, err
	}

	return user, nil
}

// FindUserByName attempts to return a row from the Users table with the matching User.Name
func FindUserByName(db *gorm.DB, name string) (*User, error) {
	user := &User{}
	err := db.First(&user, "name = ?", name).Error
	if err != nil {
		return &User{}, err
	}

	return user, nil
}
