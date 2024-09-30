package token

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(g.RandomString(32))
	assert.NoError(t, err)

	username := g.RandomUsername()
	duration := time.Minute
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(duration)
	notBefore := time.Now()

	pasetoToken, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, pasetoToken)

	payload, err := maker.VerifyToken(pasetoToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)

	assert.NotZero(t, payload.ID)
	assert.Equal(t, payload.Username, username)
	assert.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	assert.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Second)
	assert.WithinDuration(t, notBefore, payload.NotBefore, time.Second)
}

func TestExpiredPasetoToken(t *testing.T) {
	maker, err := NewPasetoMaker(g.RandomString(32))
	assert.NoError(t, err)

	username := g.RandomUsername()

	pasetoToken, err := maker.CreateToken(username, -time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, pasetoToken)

	payload, err := maker.VerifyToken(pasetoToken)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrTokenExpired.Error())
	assert.Nil(t, payload)
}

func TestInvalidPasetoToken(t *testing.T) {
	payload, err := NewPayload(g.RandomUsername(), time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)

	maker, err := NewPasetoMaker(g.RandomString(32))
	assert.NoError(t, err)

	pasetoToken, err := maker.CreateToken(g.RandomUsername(), time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, pasetoToken)

	maker, err = NewPasetoMaker(g.RandomString(32))
	assert.NoError(t, err)

	payload, err = maker.VerifyToken(pasetoToken)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrInvalidToken.Error())
	assert.Nil(t, payload)
}
