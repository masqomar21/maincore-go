package utilities

import (
	"errors"
	"time"

	"maincore_go/config"

	"github.com/golang-jwt/jwt/v5"
)

type JwtPayload struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"`
	RoleType string `json:"roleType"`
	Purpose  string `json:"purpose"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(payload JwtPayload, expiresIn time.Duration) (string, error) {
	payload.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func VerifyAccessToken(tokenString string) (*JwtPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtPayload{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtPayload); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
