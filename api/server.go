package api

import (
	"encoding/base32"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/token"
	"learn.bleckshiba/banking/util"
)

type Server struct {
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	config     util.Config
}

type PaginationParam struct {
	Page int32 `form:"page" binding:"required,min=1"`
	Size int32 `form:"size" binding:"required,min=5,max=50"`
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(base32.HexEncoding.EncodeToString([]byte(config.Paseto.SymmetricToken)))
	if err != nil {
		return nil, fmt.Errorf("unable to spawn token generator: %w", err)
	}
	server := &Server{store: store, tokenMaker: tokenMaker, config: config}
	//if err = router.SetTrustedProxies([]string{"192.168.33.121"}); err != nil {
	//	return nil, err
	//}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err = v.RegisterValidation("currency", validCurrency); err != nil {
			return nil, err
		}
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.POST("/api/v1/user", server.createUser)
	router.POST("/api/v1/login", server.loginUser)

	router.POST("/api/v1/accounts", server.createAccount)
	router.GET("/api/v1/accounts", server.getAccounts)
	router.GET("/api/v1/accounts/:id", server.getAccount)

	router.POST("/api/v1/transfer", server.createNewTransfer)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
