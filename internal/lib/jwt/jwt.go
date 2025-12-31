package jwt

import (
	"sso/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString(app.SecretHash)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//func PendingToken(user models.User, app models.App, duration time.Duration) (string, error) {
//	claims := jwt.MapClaims{
//		"uid": user.ID,
//		"exp": time.Now().Add(duration).Unix(),
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString([]byte(app.Secret))
//}
