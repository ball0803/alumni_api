package auth

import (
	"alumni_api/config"
	"alumni_api/internal/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(config.GetEnv("JWT_SECRET_KEY", "secret_key"))

func ExtractJWT(c *fiber.Ctx) (string, bool) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	return authHeader[len("Bearer "):], true
}

func ExtractJWT_Cookie(c *fiber.Ctx) (string, bool) {
	tokenString := c.Cookies("jwt")
	if tokenString == "" {
		return "", false
	}

	return tokenString, true
}

// GenerateJWT generates a JWT for a user
func GenerateJWT(userID, role string, admitYear int) (string, error) {
	claims := models.Claims{
		UserID:    userID,
		Role:      role,
		AdmitYear: admitYear,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJWT validates and parses a JWT
func ParseJWT(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func GenerateVerificationToken() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		fmt.Printf("Failed to generate token: %v", err)
	}
	return hex.EncodeToString(token)
}

func GenerateRefNum() string {
	token := make([]byte, 6)
	_, err := rand.Read(token)
	if err != nil {
		fmt.Printf("Failed to generate ref: %v", err)
	}
	return hex.EncodeToString(token)
}

func GenerateOneTimeRegistryJWT(email string) (string, error) {
	OTR := models.OneTimeRegistryJWT{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, OTR)
	return token.SignedString(jwtSecret)
}

func ParseOTRJWT(tokenString string) (*models.OneTimeRegistryJWT, error) {
	claims := &models.OneTimeRegistryJWT{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func GenerateVerificationJWT(userID, verifyToken string) (string, error) {
	verify := models.Verify{
		UserID:            userID,
		VerificationToken: verifyToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, verify)
	return token.SignedString(jwtSecret)
}

func ParseVerification(tokenString string) (*models.Verify, error) {
	claims := &models.Verify{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

// HashPassword hashes a plain password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hashed version
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
