package models

type User struct {
	ID        int64
	Title     string
	BirthDate string
	Name      string
	LastName  string
	Email     string
	PassHash  []byte
	Phone     string
}
