package jwtmanager

import (
	"crypto/rand"
	"fmt"
	"time"

	models "github.com/AleGaliev/gofermart/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	key []byte
}

func NewJWTManager(key string) *JWTManager {
	manager := &JWTManager{}
	manager.generateSecretKey(key)
	return manager
}

func (j *JWTManager) generateSecretKey(key string) error {
	if key == "" {
		j.key = make([]byte, 32)
		_, err := rand.Read(j.key)
		if err != nil {
			return err
		}
		return nil
	}

	j.key = []byte(key)
	return nil
}

func (j *JWTManager) IssueJWT(user models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.Login,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "gophermart",
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(j.key)
}

func (j *JWTManager) ValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return j.key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := jwtToken.Claims.(*jwt.RegisteredClaims)
	if !ok || !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (j *JWTManager) GetLoginFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}
