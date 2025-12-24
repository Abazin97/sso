package auth

import (
	"sso/internal/domain/models"

	ssov1 "github.com/Abazin97/protos/gen/go/sso"
)

func ToProtoUser(u models.User) *ssov1.User {
	return &ssov1.User{
		Title:     u.Title,
		Name:      u.Name,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		BirthDate: u.BirthDate,
	}
}
