package jwttoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidKey   = errors.New("invalid authorization key")
	ErrTokenExpired = errors.New("auth token expired")
)

type TokenMakerInterface interface {
	CreateToken(username string) (string, error)
	VerifyToken(tokenString string) (string, error)
}

type TokenMaker struct {
	secretKey []byte
	ttl       time.Duration
}

func CreateTokenMaker(secretKey []byte, ttl time.Duration) TokenMakerInterface {
	return &TokenMaker{secretKey, ttl}
}

func (tm *TokenMaker) CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(tm.ttl).Unix(),
		})

	return token.SignedString(tm.secretKey)
}

func (tm *TokenMaker) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (interface{}, error) {
		return tm.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrTokenExpired
		}
		return "", ErrInvalidKey
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["username"].(string), nil
	}

	return "", ErrInvalidKey
}
