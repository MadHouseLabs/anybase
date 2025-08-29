package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/madhouselabs/anybase/internal/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateTokenPair(userID primitive.ObjectID, email string, roles, permissions []string) (string, string, error)
	GenerateAccessToken(userID primitive.ObjectID, email string, roles, permissions []string) (string, error)
	GenerateRefreshToken(userID primitive.ObjectID, email string) (string, error)
	ValidateToken(tokenString string, tokenType TokenType) (*Claims, error)
	RefreshAccessToken(refreshToken string) (string, error)
}

type tokenService struct {
	config *config.AuthConfig
	secret []byte
}

func NewTokenService(cfg *config.AuthConfig) TokenService {
	return &tokenService{
		config: cfg,
		secret: []byte(cfg.JWTSecret),
	}
}

func (s *tokenService) GenerateTokenPair(userID primitive.ObjectID, email string, roles, permissions []string) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(userID, email, roles, permissions)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(userID, email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *tokenService) GenerateAccessToken(userID primitive.ObjectID, email string, roles, permissions []string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID.Hex(),
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "anybase",
			Subject:   userID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *tokenService) GenerateRefreshToken(userID primitive.ObjectID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID.Hex(),
		Email:     email,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "anybase",
			Subject:   userID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

func (s *tokenService) ValidateToken(tokenString string, tokenType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != tokenType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", tokenType, claims.TokenType)
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

func (s *tokenService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Note: In a real implementation, you would fetch the latest user data
	// including roles and permissions from the database here
	// For now, we'll just generate a new access token with the existing claims
	accessToken, err := s.GenerateAccessToken(userID, claims.Email, claims.Roles, claims.Permissions)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}