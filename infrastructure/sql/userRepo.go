package sql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kroticw/freshman-server/internal/users"
)

type UserRepo struct {
	pool *pgxpool.Pool
	tx   *pgxpool.Tx
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetById(ctx context.Context, id uint64) (*users.User, error) {
	var err error
	var rows pgx.Rows
	query := "SELECT * FROM \"user\" WHERE id = $1"
	if r.tx != nil {
		rows, err = r.tx.Query(ctx, query, id)
	} else {
		rows, err = r.pool.Query(ctx, query, id)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[users.User])
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) Save(ctx context.Context, user *users.User) error {
	if user.ID == 0 {
		return r.Create(ctx, user)
	}
	return r.Update(ctx, user)
}

func (r *UserRepo) Create(ctx context.Context, u *users.User) error {
	query := `
INSERT INTO users (name, email, password_hash, salt) VALUES ($1, $2, $3, $4) RETURNING id
`
	var row pgx.Row
	if r.tx != nil {
		row = r.tx.QueryRow(
			ctx,
			query,
			u.Name,
			u.Email,
			u.PasswordHash,
			u.Salt,
		)
	} else {
		row = r.pool.QueryRow(
			ctx,
			query,
			u.Name,
			u.Email,
			u.PasswordHash,
			u.Salt,
		)
	}
	var result uint64
	if err := row.Scan(&result); err != nil {
		return err
	}
	u.ID = result
	return nil
}

func (r *UserRepo) Update(ctx context.Context, u *users.User) error {
	query := `
UPDATE users 
SET name = $1, 
    email = $2, 
    password_hash = $3, 
    salt = $4 
WHERE id = $5
`
	var err error
	if r.tx != nil {
		_, err = r.tx.Exec(
			ctx,
			query,
			u.Name,
			u.Email,
			u.PasswordHash,
			u.Salt,
			u.ID,
		)
	} else {
		_, err = r.pool.Exec(
			ctx,
			query,
			u.Name,
			u.Email,
			u.PasswordHash,
			u.Salt,
			u.ID,
		)
	}

	return err
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {

}
