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
}

type PaginationParam struct {
	Page int32 `form:"page" binding:"required,min=1"`
	Size int32 `form:"size" binding:"required,min=5,max=50"`
}

func NewServer(store db.Store) (*Server, error) {
	config, err := util.LoadConfig("../.")
	tokenMaker, err := token.NewPasetoMaker(base32.HexEncoding.EncodeToString([]byte(config.Paseto.SymmetricToken)))
	if err != nil {
		return nil, fmt.Errorf("unable to spawn token generator: %w", err)
	}
	server := &Server{store: store, tokenMaker: tokenMaker}
	router := gin.Default()
	//if err = router.SetTrustedProxies([]string{"192.168.33.121"}); err != nil {
	//	return nil, err
	//}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err = v.RegisterValidation("currency", validCurrency); err != nil {
			return nil, err
		}
	}

	router.POST("/user", server.createUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts", server.getAccounts)
	router.GET("/accounts/:id", server.getAccount)

	router.POST("/transfer", server.createNewTransfer)

	server.router = router
	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
