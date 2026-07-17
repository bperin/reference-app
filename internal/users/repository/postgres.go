package repository

import (
	"context"
	"errors"

	"github.com/example/reference-app/internal/users"
	userssqlc "github.com/example/reference-app/internal/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostgresRepository struct {
	db *userssqlc.Queries
}

func NewPostgres(db *userssqlc.Queries) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, u *users.User) error {
	row, err := r.db.CreateUser(ctx, userssqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: u.ID, Valid: true},
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		CreatedAt:    pgtype.Timestamptz{Time: u.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
	})
	if err != nil {
		return err
	}
	u.ID = row.ID.Bytes
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	row, err := r.db.GetUserByID(ctx, userssqlc.GetUserByIDParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrNotFound
		}
		return nil, err
	}
	return &users.User{
		ID:           row.ID.Bytes,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	row, err := r.db.GetUserByEmail(ctx, userssqlc.GetUserByEmailParams{
		Email: email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrNotFound
		}
		return nil, err
	}
	return &users.User{
		ID:           row.ID.Bytes,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

func (r *PostgresRepository) Update(ctx context.Context, u *users.User) error {
	_, err := r.db.UpdateUser(ctx, userssqlc.UpdateUserParams{
		ID:          pgtype.UUID{Bytes: u.ID, Valid: true},
		DisplayName: u.DisplayName,
		UpdatedAt:   pgtype.Timestamptz{Time: u.UpdatedAt, Valid: true},
	})
	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.DeleteUser(ctx, userssqlc.DeleteUserParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
	})
}
