package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/user"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed login attempts")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrEmailNotVerified   = errors.New("email address is not verified")
)

type Service interface {
	Register(ctx context.Context, req *models.UserRegistration) (*models.User, error)
	Login(ctx context.Context, req *models.UserLogin) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID primitive.ObjectID) error
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error
}

type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type service struct {
	userRepo     user.Repository
	tokenService TokenService
	config       *config.AuthConfig
}

func NewService(userRepo user.Repository, config *config.AuthConfig) Service {
	return &service{
		userRepo:     userRepo,
		tokenService: NewTokenService(config),
		config:       config,
	}
}

func (s *service) Register(ctx context.Context, req *models.UserRegistration) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, user.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate email verification token
	verificationToken := s.generateToken()

	// Create user
	newUser := &models.User{
		Email:                  req.Email,
		Username:               req.Username,
		Password:               hashedPassword,
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		EmailVerificationToken: verificationToken,
		Role:                   "developer", // Default role for new users
		UserType:               models.UserTypeRegular,
		Active:                 true,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Send verification email with token

	// Clear sensitive fields before returning
	newUser.Password = ""
	newUser.EmailVerificationToken = ""

	return newUser, nil
}

func (s *service) Login(ctx context.Context, req *models.UserLogin) (*AuthResponse, error) {
	// Get user by email
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is locked
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	// Check if account is active
	if !u.Active {
		return nil, ErrAccountInactive
	}

	// Verify password
	if err := s.verifyPassword(u.Password, req.Password); err != nil {
		// Increment login attempts
		attempts := u.LoginAttempts + 1
		if err := s.userRepo.UpdateLoginAttempts(ctx, u.ID, attempts); err != nil {
			// Log error but don't fail the login attempt
			fmt.Printf("failed to update login attempts: %v\n", err)
		}

		// Lock account if max attempts reached
		if attempts >= s.config.MaxLoginAttempts {
			lockUntil := time.Now().Add(s.config.LockoutDuration)
			if err := s.userRepo.UpdateLockedUntil(ctx, u.ID, &lockUntil); err != nil {
				fmt.Printf("failed to lock account: %v\n", err)
			}
			return nil, ErrAccountLocked
		}

		return nil, ErrInvalidCredentials
	}

	// Reset login attempts and update last login
	if err := s.userRepo.UpdateLastLogin(ctx, u.ID); err != nil {
		fmt.Printf("failed to update last login: %v\n", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.tokenService.GenerateTokenPair(
		u.ID,
		u.Email,
		[]string{u.Role}, // Convert single role to array for token
		[]string{}, // Permissions come from roles now
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Clear sensitive fields
	u.Password = ""

	return &AuthResponse{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := s.tokenService.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user ID from claims
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Get updated user data
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is active
	if !u.Active {
		return nil, ErrAccountInactive
	}

	// Generate new access token with updated permissions
	accessToken, err := s.tokenService.GenerateAccessToken(
		u.ID,
		u.Email,
		[]string{u.Role}, // Convert single role to array for token
		[]string{}, // Permissions come from roles now
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Clear sensitive fields
	u.Password = ""

	return &AuthResponse{
		User:         u,
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return the same refresh token
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
	}, nil
}

func (s *service) Logout(ctx context.Context, userID primitive.ObjectID) error {
	// TODO: Implement token blacklist or session management
	// For now, logout is handled on the client side
	return nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Get user by verification token
	// For now, this is a placeholder
	return fmt.Errorf("email verification not yet implemented")
}

func (s *service) RequestPasswordReset(ctx context.Context, email string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken := s.generateToken()
	expiry := time.Now().Add(1 * time.Hour)

	if err := s.userRepo.SetPasswordResetToken(ctx, u.ID, resetToken, expiry); err != nil {
		return fmt.Errorf("failed to set reset token: %w", err)
	}

	// TODO: Send password reset email with token

	return nil
}

func (s *service) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Get user by reset token
	u, err := s.userRepo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, u.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error {
	// Get user
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := s.verifyPassword(u.Password, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *service) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *service) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *service) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}