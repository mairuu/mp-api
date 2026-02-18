package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mairuu/mp-api/internal/features/user/model"
	repo "github.com/mairuu/mp-api/internal/features/user/repository"
	"github.com/mairuu/mp-api/internal/platform/authorization"
)

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, role string) (string, error)
}

type Service struct {
	repo           repo.Repository
	tokenGenerator TokenGenerator
	enforcer       *authorization.Enforcer
}

func NewService(repo repo.Repository, tokenGenerator TokenGenerator, enforcer *authorization.Enforcer) *Service {
	return &Service{
		repo:           repo,
		tokenGenerator: tokenGenerator,
		enforcer:       enforcer,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterDTO) (*UserResponseDTO, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u, err := model.NewUser(req.Username, req.Email, string(passwordHash))
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveUser(ctx, u); err != nil {
		return nil, err
	}

	return &UserResponseDTO{
		ID:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role.String(),
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginDTO) (*LoginResponseDTO, error) {
	u, err := s.repo.GetUserByEmailOrUsername(ctx, req.EmailOrUsername)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrInvalidCredentials
		}
		return nil, err
	}

	// verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	// generate JWT token
	token, err := s.tokenGenerator.GenerateToken(u.ID, u.Role.String())
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponseDTO{
		User: UserResponseDTO{
			ID:       u.ID.String(),
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role.String(),
		},
		Token: token,
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResponseDTO, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &UserResponseDTO{
		ID:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
	}, nil
}
