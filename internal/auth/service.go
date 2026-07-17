package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/example/reference-app/internal/users"
)

type PasswordVerifier interface {
	Verify(hashedPassword, password string) error
	Hash(password string) (string, error)
}

type AccessTokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, scopes []string, roles []string) (string, error)
}

type RefreshTokenGenerator interface {
	Generate() (string, error)
	Hash(token string) string
}

type Service struct {
	userRepo        users.Repository
	refreshRepo     RefreshRepository
	passwords       PasswordVerifier
	accessTokens    AccessTokenIssuer
	refreshTokens   RefreshTokenGenerator
	logger          *slog.Logger
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(
	userRepo users.Repository,
	refreshRepo RefreshRepository,
	passwords PasswordVerifier,
	accessTokens AccessTokenIssuer,
	refreshTokens RefreshTokenGenerator,
	logger *slog.Logger,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		userRepo:        userRepo,
		refreshRepo:     refreshRepo,
		passwords:       passwords,
		accessTokens:    accessTokens,
		refreshTokens:   refreshTokens,
		logger:          logger.With(slog.String("component", "auth_service")),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
}

func (s *Service) RegisterUser(ctx context.Context, email, password, displayName string) (*users.User, error) {
	if email == "" || password == "" || displayName == "" {
		return nil, users.ErrInvalidInput
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	now := time.Now()
	user := &users.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  displayName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "user registered successfully", slog.String("user_id", user.ID.String()))
	return user, nil
}

func (s *Service) ExchangePassword(ctx context.Context, email, password string, scope []string) (*TokenResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	if err := s.passwords.Verify(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Issue bearer access token
	accessToken, err := s.accessTokens.IssueAccessToken(u.ID, scope, []string{"user"})
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	// Issue rotating refresh token
	rawRefreshToken, err := s.refreshTokens.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := s.refreshTokens.Hash(rawRefreshToken)
	expiresAt := time.Now().Add(s.refreshTokenTTL)

	session := &RefreshSession{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: tokenHash,
		FamilyID:  uuid.New(), // brand new token family
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.refreshRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create refresh session: %w", err)
	}

	s.logger.InfoContext(ctx, "issued oauth tokens successfully",
		slog.String("user_id", u.ID.String()),
		slog.String("family_id", session.FamilyID.String()),
	)

	return &TokenResponse{
		AccessToken:           accessToken,
		TokenType:             "Bearer",
		ExpiresIn:             int64(s.accessTokenTTL.Seconds()),
		RefreshToken:          rawRefreshToken,
		RefreshTokenExpiresIn: int64(s.refreshTokenTTL.Seconds()),
		Scope:                 "user",
	}, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*TokenResponse, error) {
	tokenHash := s.refreshTokens.Hash(rawToken)

	session, err := s.refreshRepo.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshSessionNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("lookup refresh session: %w", err)
	}

	// Detect token reuse
	if session.IsRevoked() {
		s.logger.WarnContext(ctx, "refresh token reuse detected! revoking entire family chain",
			slog.String("user_id", session.UserID.String()),
			slog.String("family_id", session.FamilyID.String()),
		)
		// Revoke the entire token family
		_ = s.refreshRepo.RevokeFamily(ctx, session.FamilyID)
		return nil, ErrRefreshTokenRevoked
	}

	// Validate expiration
	if session.IsExpired() {
		return nil, ErrRefreshTokenExpired
	}

	// Valid session -> generate a replacement before atomically consuming the
	// current session and persisting the replacement.
	newRawToken, err := s.refreshTokens.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate replacement refresh token: %w", err)
	}

	newHash := s.refreshTokens.Hash(newRawToken)
	newSession := &RefreshSession{
		ID:        uuid.New(),
		UserID:    session.UserID,
		TokenHash: newHash,
		FamilyID:  session.FamilyID, // reuse family ID
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
		CreatedAt: time.Now(),
	}

	accessToken, err := s.accessTokens.IssueAccessToken(session.UserID, []string{"user"}, []string{"user"})
	if err != nil {
		return nil, fmt.Errorf("issue rotated access token: %w", err)
	}

	if err := s.refreshRepo.RotateSession(ctx, session.ID, newSession); err != nil {
		if errors.Is(err, ErrRefreshSessionNotFound) {
			_ = s.refreshRepo.RevokeFamily(ctx, session.FamilyID)
			return nil, ErrRefreshTokenRevoked
		}
		return nil, fmt.Errorf("create rotated refresh session: %w", err)
	}

	s.logger.InfoContext(ctx, "rotated refresh token successfully",
		slog.String("user_id", session.UserID.String()),
		slog.String("family_id", session.FamilyID.String()),
	)

	return &TokenResponse{
		AccessToken:           accessToken,
		TokenType:             "Bearer",
		ExpiresIn:             int64(s.accessTokenTTL.Seconds()),
		RefreshToken:          newRawToken,
		RefreshTokenExpiresIn: int64(s.refreshTokenTTL.Seconds()),
		Scope:                 "user",
	}, nil
}

func (s *Service) Revoke(ctx context.Context, rawToken string) error {
	tokenHash := s.refreshTokens.Hash(rawToken)

	session, err := s.refreshRepo.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshSessionNotFound) {
			return ErrInvalidRefreshToken
		}
		return fmt.Errorf("lookup session for revocation: %w", err)
	}

	if err := s.refreshRepo.RevokeSession(ctx, session.ID); err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}

	s.logger.InfoContext(ctx, "refresh token revoked",
		slog.String("user_id", session.UserID.String()),
		slog.String("family_id", session.FamilyID.String()),
	)

	return nil
}
