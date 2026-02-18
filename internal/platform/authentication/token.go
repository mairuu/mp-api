package authentication

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	secret         []byte
	accessTokenTTL time.Duration
}

func NewTokenService(secret []byte, accessTokenTTL time.Duration) *TokenService {
	return &TokenService{
		secret:         secret,
		accessTokenTTL: accessTokenTTL,
	}
}

func (s *TokenService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *TokenService) ValidateToken(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, "", fmt.Errorf("invalid user_id in token")
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, "", fmt.Errorf("parse user_id: %w", err)
		}
		role, ok := claims["role"].(string)
		if !ok {
			return uuid.Nil, "", fmt.Errorf("invalid role in token")
		}
		return userID, role, nil
	}

	return uuid.Nil, "", fmt.Errorf("invalid token")
}
