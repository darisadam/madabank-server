package service

import (
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/darisadam/madabank-server/internal/pkg/crypto"
	"github.com/darisadam/madabank-server/internal/pkg/jwt"
	"github.com/darisadam/madabank-server/internal/repository"
	"github.com/google/uuid"
)

type UserService interface {
	Register(req *user.CreateUserRequest) (*user.User, error)
	Login(req *user.LoginRequest) (*user.LoginResponse, error)
	GetProfile(userID uuid.UUID) (*user.User, error)
	UpdateProfile(userID uuid.UUID, req *user.UpdateUserRequest) (*user.User, error)
	DeleteAccount(userID uuid.UUID) error
}

type userService struct {
	userRepo   repository.UserRepository
	jwtService *jwt.JWTService
}

func NewUserService(userRepo repository.UserRepository, jwtService *jwt.JWTService) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *userService) Register(req *user.CreateUserRequest) (*user.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Parse date of birth if provided
	var dob *time.Time
	if req.DateOfBirth != nil {
		parsedDOB, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date of birth format, use YYYY-MM-DD")
		}
		dob = &parsedDOB
	}

	// Create user
	newUser := &user.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		DateOfBirth:  dob,
		KYCStatus:    "pending",
		IsActive:     true,
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove sensitive data before returning
	newUser.PasswordHash = ""

	return newUser, nil
}

func (s *userService) Login(req *user.LoginRequest) (*user.LoginResponse, error) {
	// Get user by email
	u, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is active
	if !u.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if !crypto.CheckPassword(req.Password, u.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtService.GenerateToken(u.ID, u.Email, "customer")
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Remove sensitive data
	u.PasswordHash = ""

	return &user.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      u,
	}, nil
}

func (s *userService) GetProfile(userID uuid.UUID) (*user.User, error) {
	u, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	u.PasswordHash = ""
	return u, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, req *user.UpdateUserRequest) (*user.User, error) {
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.DateOfBirth != nil {
		parsedDOB, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date of birth format")
		}
		updates["date_of_birth"] = parsedDOB
	}

	if len(updates) == 0 {
		return s.GetProfile(userID)
	}

	if err := s.userRepo.Update(userID, updates); err != nil {
		return nil, err
	}

	return s.GetProfile(userID)
}

func (s *userService) DeleteAccount(userID uuid.UUID) error {
	return s.userRepo.Delete(userID)
}
