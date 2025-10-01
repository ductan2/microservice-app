package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword compares plain password with a bcrypt hash.
func CheckPassword(hash string, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}
