package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type UserTokenHandler struct {
	secretKey string
}

func NewUserTokenHandler(secretKey string) *UserTokenHandler {
	return &UserTokenHandler{
		secretKey: secretKey,
	}
}

func (u *UserTokenHandler) Generate(email string, idempotencyKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["idempotencyKey"] = idempotencyKey
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(u.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}

func (u *UserTokenHandler) GenerateToken(email string, idempotencyKey string) string {
	token, err := u.Generate(email, idempotencyKey)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return ""
	}
	return token
}
