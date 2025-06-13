package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func GenerateJWT(id primitive.ObjectID, role string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id.Hex(),
		"role": role,
		"exp":  time.Now().Add(12 * time.Hour).Unix(), // Hết hạn sau 12 giờ
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}
