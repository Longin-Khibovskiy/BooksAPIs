package models

type User struct {
	ID           int
	Email        string
	PasswordHash string
	Name         string
	Role         string
	AvatarURL    string
	CreatedAt    string
}
