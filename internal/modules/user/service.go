package user

import (
	"context"
	"errors"

	"github.com/sabih15/TeleOpServer/internal/auth"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service is the interface — equivalent to IUserService in .NET.
type IService interface {
	Register(ctx context.Context, req RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Delete(ctx context.Context, userID uint) error
	GetProfile(ctx context.Context, userID uint) (*UserResponse, error)
}

type service struct {
	repo IRepository
	cfg  *config.Config
}

// NewService is the constructor — Wire injects Repository and *config.Config.
func NewService(repo IRepository, cfg *config.Config) IService {
	return &service{repo: repo, cfg: cfg}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return toResponse(u), nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(u.ID, s.cfg.JWT.Secret, s.cfg.JWT.ExpiryHours)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}

func (s *service) Delete(ctx context.Context, userID uint) error {
	return s.repo.Delete(ctx, userID)
}

func (s *service) GetProfile(ctx context.Context, userID uint) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toResponse(u), nil
}

func toResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
	}
}
