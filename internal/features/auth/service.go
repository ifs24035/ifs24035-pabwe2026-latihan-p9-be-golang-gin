package auth

import (
	"errors"
	"time"

	"app/internal/database"
	"app/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	GetProfile(userID uint) (*UserResponse, error)
	UpdateProfile(userID uint, req UpdateProfileRequest) (*UserResponse, error)
}

type authService struct {
	repo      UserRepository
	jwtSecret string
}

func NewAuthService(repo UserRepository, jwtSecret string) AuthService {
	return &authService{repo: repo, jwtSecret: jwtSecret}
}

func (s *authService) Register(req RegisterRequest) (*AuthResponse, error) {
	existing, _ := s.repo.FindByUsernameOrEmail(req.Username)
	if existing != nil {
		return nil, errors.New("username already registered")
	}
	existingEmail, _ := s.repo.FindByUsernameOrEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	role := req.Role
	if role == "" || (role != "admin" && role != "member") {
		role = "member"
	}

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	_ = s.repo.Create(user)
	token, _ := s.generateToken(user)

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *authService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByUsernameOrEmail(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, _ := s.generateToken(user)

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *authService) GetProfile(userID uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *authService) UpdateProfile(userID uint, req UpdateProfileRequest) (*UserResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.Password = string(hashed)
	}

	_ = s.repo.Update(user)

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *authService) generateToken(user *database.User) (string, error) {
	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
