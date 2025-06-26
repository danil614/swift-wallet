package repository

import (
	"context"
	"errors"
	"time"

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

const maxRetry = 5

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
		bal int64
		err error
	)

	for i := 0; i < maxRetry; i++ {
		bal, err = r.changeOnce(ctx, id, op, amount)
		if err != nil && isSerializationErr(err) {
			time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			continue
		}
		break
	}
	return bal, err
}

func (r *repo) changeOnce(ctx context.Context, id uuid.UUID, op model.OperationType, amount int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// блокируем строку
	var current int64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id=$1 FOR UPDATE`, id).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	var newBal int64
	switch op {
	case model.Deposit:
		newBal = current + amount
	case model.Withdraw:
		if current < amount {
			return 0, ErrInsufficientFunds
		}
		newBal = current - amount
	default:
		return 0, ErrUnknownOperation
	}

	if _, err = tx.Exec(ctx,
		`UPDATE wallets SET balance=$1, updated_at=now() WHERE id=$2`, newBal, id); err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return newBal, nil
}

func isSerializationErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "40001"
	}
	return false
}
