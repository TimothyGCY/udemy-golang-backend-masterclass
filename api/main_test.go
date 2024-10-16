package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/util"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	pasetoConfig := util.Paseto{
		SymmetricToken:     "abcdefghijklmnop",
		AccessTokenTimeout: time.Minute,
	}
	config := util.Config{
		Paseto: pasetoConfig,
	}

	server, err := NewServer(config, store)
	assert.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
