package model

type UserModel struct {
	ID int
	Username string
	PasswordHash string
	Storage int64
}