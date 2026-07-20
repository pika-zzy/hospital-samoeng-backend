package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID uint
	Role   string
	// TokenVersion — snapshot ของ User.TokenVersion ตอนออก token (CRIT-01)
	// middleware เทียบค่านี้กับค่าปัจจุบันใน DB — ไม่ตรง = token ถูกเพิกถอนแล้ว
	TokenVersion int
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, role string, tokenVersion int) (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
