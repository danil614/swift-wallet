package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"swiftwallet/internal/model"
)

var (
	ErrNotFound          = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrUnknownOperation  = errors.New("unknown operation")
)

type Repository interface {
	GetBalance(ctx context.Context, id uuid.UUID) (int64, error)
	ChangeBalance(ctx context.Context, id uuid.UUID, op model.OperationType, amount int64) (int64, error)
}

type PgxIface interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type repo struct{ db PgxIface }

func New(db PgxIface) Repository { return &repo{db: db} }

func (r *repo) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	var bal int64
	err := r.db.QueryRow(ctx, `SELECT balance FROM wallets WHERE id=$1`, id).Scan(&bal)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return bal, err
}

func (r *repo) ChangeBalance(ctx context.Context, id uuid.UUID, op model.OperationType, amount int64) (int64, error) {
	var (
		sql string
		bal int64
	)

	switch op {
	case model.Deposit:
		sql = `UPDATE wallets
                  SET balance   = balance + $1,
                      updated_at = now()
                WHERE id = $2
            RETURNING balance`

	case model.Withdraw:
		sql = `UPDATE wallets
                  SET balance   = balance - $1,
                      updated_at = now()
                WHERE id = $2
                  AND balance  >= $1
            RETURNING balance`

	default:
		return 0, ErrUnknownOperation
	}

	err := r.db.QueryRow(ctx, sql, amount, id).Scan(&bal)
	if errors.Is(err, pgx.ErrNoRows) {
		if op == model.Withdraw {
			return 0, ErrInsufficientFunds
		}
		return 0, ErrNotFound
	}
	return bal, err
}
