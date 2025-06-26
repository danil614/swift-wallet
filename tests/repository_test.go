package tests

import (
	"context"
	"github.com/jackc/pgx/v5"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"

	"swiftwallet/internal/model"
	"swiftwallet/internal/repository"
)

func TestGetBalanceSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	want := int64(777)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT balance FROM wallets WHERE id=$1`)).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(want))

	repo := repository.New(mock)
	got, err := repo.GetBalance(context.Background(), id)

	require.NoError(t, err)
	require.Equal(t, want, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeBalanceDeposit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	amount := int64(100)
	expected := int64(150)

	mock.ExpectQuery(regexp.QuoteMeta(
		`UPDATE wallets
              SET balance   = balance + $1,
                  updated_at = now()
            WHERE id = $2
        RETURNING balance`)).
		WithArgs(amount, id).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(expected))

	repo := repository.New(mock)
	got, err := repo.ChangeBalance(context.Background(), id, model.Deposit, amount)
	require.NoError(t, err)
	require.Equal(t, expected, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeBalanceWithdrawInsufficient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	amount := int64(200)

	mock.ExpectQuery(regexp.QuoteMeta(
		`UPDATE wallets
              SET balance   = balance - $1,
                  updated_at = now()
            WHERE id = $2
              AND balance  >= $1
        RETURNING balance`)).
		WithArgs(amount, id).
		WillReturnError(pgx.ErrNoRows)

	repo := repository.New(mock)
	_, err = repo.ChangeBalance(context.Background(), id, model.Withdraw, amount)
	require.ErrorIs(t, err, repository.ErrInsufficientFunds)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBalanceNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT balance FROM wallets WHERE id=$1`)).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	repo := repository.New(mock)
	_, err = repo.GetBalance(context.Background(), id)
	require.ErrorIs(t, err, repository.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
