package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mairuu/mp-api/internal/features/user/model"
	repo "github.com/mairuu/mp-api/internal/features/user/repository"
	"github.com/mairuu/mp-api/internal/platform/authorization"
)

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, role string) (string, error)
	GenerateRefreshToken() (string, error)
	RefreshTokenTTL() time.Duration
}

type Service struct {
	repo           repo.Repository
	tokenGenerator TokenGenerator
	enforcer       *authorization.Enforcer
	cleanupTicker  *time.Ticker
	cleanupDone    chan struct{}
}

func NewService(repo repo.Repository, tokenGenerator TokenGenerator, enforcer *authorization.Enforcer) *Service {
	return &Service{
		repo:           repo,
		tokenGenerator: tokenGenerator,
		enforcer:       enforcer,
		cleanupDone:    make(chan struct{}),
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

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	accessToken, err := s.tokenGenerator.GenerateToken(u.ID, u.Role.String())
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := s.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	rt := model.NewRefreshToken(rawRefreshToken, u.ID, s.tokenGenerator.RefreshTokenTTL())
	if err := s.repo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &LoginResponseDTO{
		User: UserResponseDTO{
			ID:       u.ID.String(),
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role.String(),
		},
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenDTO) (*LoginResponseDTO, error) {
	rt, err := s.repo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	if rt.IsRevoked() {
		return nil, model.ErrRefreshTokenRevoked
	}
	if rt.IsExpired() {
		return nil, model.ErrRefreshTokenExpired
	}

	// revoke the old token (rotation)
	if err := s.repo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	u, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenGenerator.GenerateToken(u.ID, u.Role.String())
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := s.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	newRT := model.NewRefreshToken(rawRefreshToken, u.ID, s.tokenGenerator.RefreshTokenTTL())
	if err := s.repo.SaveRefreshToken(ctx, newRT); err != nil {
		return nil, fmt.Errorf("save new refresh token: %w", err)
	}

	return &LoginResponseDTO{
		User: UserResponseDTO{
			ID:       u.ID.String(),
			Username: u.Username,
			Email:    u.Email,
			Role:     u.Role.String(),
		},
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, req RefreshTokenDTO) error {
	if err := s.repo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		if errors.Is(err, model.ErrRefreshTokenNotFound) {
			return nil // idempotent
		}
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
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

func (s *Service) StartCleanup(interval time.Duration) {
	if s.cleanupTicker != nil {
		return // already running
	}
	s.cleanupTicker = time.NewTicker(interval)

	go func() {
		defer s.cleanupTicker.Stop()

		s.deleteExpiredRefreshTokens()

		for {
			select {
			case <-s.cleanupTicker.C:
				s.deleteExpiredRefreshTokens()
			case <-s.cleanupDone:
				return
			}
		}
	}()
}

func (s *Service) StopCleanup() {
	if s.cleanupDone != nil {
		close(s.cleanupDone)
	}
}

func (s *Service) deleteExpiredRefreshTokens() {
	if err := s.repo.DeleteExpiredRefreshTokens(context.Background()); err != nil {
		// non-fatal; will retry on next tick
		_ = err
	}
}
