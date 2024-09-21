package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1, acc2 := createRandomAccount(t), createRandomAccount(t)
	fmt.Println(acc1.Balance, acc2.Balance)

	n := 5
	amount := float64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParam{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		fromAcc := result.FromAccount
		require.NotEmpty(t, fromAcc)
		require.Equal(t, acc1.ID, fromAcc.ID)

		toAcc := result.ToAccount
		require.NotEmpty(t, toAcc)
		require.Equal(t, acc2.ID, toAcc.ID)
	}

	updatedAcc1, err := testQueries.GetAccountById(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance-(float64(n)*amount), updatedAcc1.Balance)

	updatedAcc2, err := testQueries.GetAccountById(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.Equal(t, acc2.Balance+(float64(n)*amount), updatedAcc2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	acc1, acc2 := createRandomAccount(t), createRandomAccount(t)

	n := 10
	amount := float64(10)
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccID, toAccID := acc1.ID, acc2.ID

		if i%2 == 1 {
			fromAccID, toAccID = toAccID, fromAccID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParam{
				FromAccountID: fromAccID,
				ToAccountID:   toAccID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAcc1, err := testQueries.GetAccountById(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAcc2, err := testQueries.GetAccountById(context.Background(), acc2.ID)
	require.NoError(t, err)

	require.Equal(t, acc1.Balance, updatedAcc1.Balance)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)
}
