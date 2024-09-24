package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"learn.bleckshiba/banking/util"
)

func createRandomUser(t *testing.T) User {
	generator := util.NewGenerator()

	hashPw, err := util.HashPassword(generator.RandomPassword())
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       generator.RandomUsername(),
		HashedPassword: hashPw,
		FullName:       generator.RandomName(),
		Email:          generator.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.ModifiedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUserByUsername(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUserByUsername(context.Background(), user1.Username)
	require.NoError(t, err)
	require.Equal(t, user1, user2)
}
