package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPassword), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Subject:   userID.String(),
	})
	token, err := newToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if !token.Valid {
		return uuid.UUID{}, errors.New("token has expired")
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is not set")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("incorrect authorization type, must be of type Bearer")
	}
	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
	return strings.TrimSpace(bearerToken), nil
}

func MakeRefreshToken() (string, error) {
	buffer := make([]byte, 32)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}
	hexData := hex.EncodeToString(buffer)
	return hexData, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is not set")
	}
	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return "", errors.New("incorrect authorization type, must be of type ApiKey")
	}
	apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
	return strings.TrimSpace(apiKey), nil
}
