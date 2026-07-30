package models

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Post struct {
	ID       int64
	Username string
	Content  string
}
