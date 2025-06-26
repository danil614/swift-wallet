package tests

import (
	"context"
	"github.com/jackc/pgx/v5"
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

	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id=\$1`).
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
	start := int64(50)
	finish := start + amount

	mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.Serializable})
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id=\$1 FOR UPDATE`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(start))
	mock.ExpectExec(`UPDATE wallets SET balance=`).
		WithArgs(finish, id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	repo := repository.New(mock)
	got, err := repo.ChangeBalance(context.Background(), id, model.Deposit, amount)
	require.NoError(t, err)
	require.Equal(t, finish, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChangeBalanceWithdrawInsufficient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	start := int64(30)
	withdraw := int64(100)

	mock.ExpectBeginTx(pgx.TxOptions{IsoLevel: pgx.Serializable})
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id=\$1 FOR UPDATE`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(start))
	// транзакция будет отменена внутри метода при ErrInsufficientFunds
	mock.ExpectRollback()

	repo := repository.New(mock)
	_, err = repo.ChangeBalance(context.Background(), id, model.Withdraw, withdraw)
	require.ErrorIs(t, err, repository.ErrInsufficientFunds)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBalanceNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	id := uuid.New()
	mock.ExpectQuery(`SELECT balance FROM wallets WHERE id=\$1`).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	repo := repository.New(mock)
	_, err = repo.GetBalance(context.Background(), id)
	require.ErrorIs(t, err, repository.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
