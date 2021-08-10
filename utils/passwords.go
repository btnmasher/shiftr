package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword takes the provided string and generates a bcrypt hash of it at the default strength
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword compares a provided string password with the provided bcrypt password hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
