package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func CheckPassword(hashedPwd, inputPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(inputPwd))
}
