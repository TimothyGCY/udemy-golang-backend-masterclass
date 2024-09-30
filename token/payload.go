package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var ErrTokenExpired = errors.New("authorization token expired")
var ErrInvalidToken = errors.New("invalid token")

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expired_at"`
	NotBefore time.Time `json:"notBefore"`
	Issuer    string    `json:"issuer"`
	Audience  []string  `json:"audience"`
}

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.ExpiresAt), nil
}

func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.IssuedAt), nil
}

func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.NotBefore), nil
}

func (p *Payload) GetIssuer() (string, error) {
	return p.Issuer, nil
}

func (p *Payload) GetSubject() (string, error) {
	return p.Username, nil
}

func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiry := now.Add(duration)
	return &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  now,
		ExpiresAt: expiry,
		NotBefore: time.Now(),
		Issuer:    "simple_bank",
		Audience:  []string{},
	}, nil
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiresAt) {
		return ErrTokenExpired
	}
	return nil
}
