package users

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.With(slog.String("component", "users_service")),
	}
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, displayName string) (*User, error) {
	if displayName == "" {
		return nil, ErrInvalidInput
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user for update: %w", err)
	}

	u.DisplayName = displayName
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	s.logger.InfoContext(ctx, "user profile updated", slog.String("user_id", id.String()))
	return u, nil
}

func (s *Service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user account: %w", err)
	}

	s.logger.InfoContext(ctx, "user account deleted", slog.String("user_id", id.String()))
	return nil
}
