package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	generator := NewGenerator()
	sourcePw := generator.RandomPassword()
	hashedPw, err := HashPassword(sourcePw)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPw)

	err = CheckPassword(sourcePw, hashedPw)
	require.NoError(t, err)

	wrongPw := generator.RandomPassword()
	err = CheckPassword(wrongPw, hashedPw)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}
