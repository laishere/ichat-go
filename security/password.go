package security

import (
	"golang.org/x/crypto/bcrypt"
)

func EncodePassword(password string) string {
	encoded, err := bcrypt.GenerateFromPassword([]byte(password), 6)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func ComparePassword(encoded, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
	return err == nil
}
