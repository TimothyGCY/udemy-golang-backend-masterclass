package token

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"learn.bleckshiba/banking/util"
	"testing"
	"time"
)

var g = util.NewGenerator()

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(g.RandomString(32))
	assert.NoError(t, err)

	username := g.RandomUsername()
	duration := time.Minute
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(duration * 10)
	notBefore := time.Now()

	token, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)

	assert.NotZero(t, payload.ID)
	assert.Equal(t, payload.Username, username)
	assert.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	assert.WithinDuration(t, expiresAt, payload.ExpiresAt, time.Minute*10)
	assert.WithinDuration(t, notBefore, payload.NotBefore, time.Second)
}

func TestExpireToken(t *testing.T) {
	maker, err := NewJWTMaker(g.RandomString(32))
	assert.NoError(t, err)

	username := g.RandomUsername()
	token, err := maker.CreateToken(username, -time.Second)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrTokenExpired.Error())
	assert.Nil(t, payload)
}

func TestNoAlg(t *testing.T) {
	payload, err := NewPayload(g.RandomUsername(), time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	maker, err := NewJWTMaker(g.RandomString(32))
	assert.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrInvalidToken.Error())
	assert.Nil(t, payload)
}
