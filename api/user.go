package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPw, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, err := server.store.CreateUser(ctx, db.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: hashedPw,
		FullName:       req.FullName,
	})

	if err == nil {
		res := userResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
		}

		ctx.JSON(http.StatusOK, res)
		return
	}

	if err, ok := err.(*pq.Error); ok {
		log.Println(err.Code.Name())
		switch err.Code.Name() {
		case "unique_violation":
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
	}
	ctx.JSON(http.StatusInternalServerError, errorResponse(err))
}
