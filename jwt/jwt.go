package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"ichat-go/config"
	"time"
)

func GenerateToken(payload string, expiration time.Time) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["pl"] = payload
	claims["iss"] = config.App.Jwt.Issuer
	claims["exp"] = expiration.Unix()
	tokenStr, err := token.SignedString([]byte(config.App.Jwt.Secret))
	if err != nil {
		panic(err)
	}
	return tokenStr
}

func ValidateToken(token string) (string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.App.Jwt.Secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		return claims["pl"].(string), nil
	}
	return "", errors.New("invalid token")
}
