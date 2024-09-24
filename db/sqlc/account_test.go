package db

import (
	"context"
	"database/sql"
	"learn.bleckshiba/banking/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	generator := util.NewGenerator()

	user := createRandomUser(t)

	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  generator.RandomMoney(),
		Currency: generator.RandomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	return acc
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccountById(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2, err := testQueries.GetAccountById(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.Equal(t, acc1, acc2)
}

func TestListAccounts(t *testing.T) {
	arg := GetAccountsParams{
		Limit:  10,
		Offset: 1,
	}
	acc, err := testQueries.GetAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, acc)
	require.LessOrEqual(t, len(acc), int(arg.Limit))
}

func TestUpdateAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	arg := UpdateBalanceParams{
		ID:      acc1.ID,
		Balance: acc1.Balance - 20,
	}
	err := testQueries.UpdateBalance(context.Background(), arg)
	require.NoError(t, err)

	acc2, _ := testQueries.GetAccountById(context.Background(), acc1.ID)
	require.Equal(t, arg.Balance, acc2.Balance)
}

func TestDeleteAccount(t *testing.T) {
	acc := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), acc.ID)
	require.NoError(t, err)

	acc2, err := testQueries.GetAccountById(context.Background(), acc.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, acc2)
}
