package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "learn.bleckshiba/banking/db/sqlc"
)

type newTransferRequest struct {
	FromAccountID int64   `json:"fromAccount" binding:"required,min=1"`
	ToAccountID   int64   `json:"toAccount" binding:"required,min=1"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,currency"`
}

func (server *Server) createNewTransfer(ctx *gin.Context) {
	var request newTransferRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if request.FromAccountID == request.ToAccountID {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("Invalid transaction accounts"))
		return
	}

	if !server.validateAccount(ctx, request.FromAccountID, request.Currency) {
		return
	}

	if !server.validateAccount(ctx, request.ToAccountID, request.Currency) {
		return
	}

	arg := db.TransferTxParam{
		FromAccountID: request.FromAccountID,
		ToAccountID:   request.ToAccountID,
		Amount:        request.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validateAccount(ctx *gin.Context, accountID int64, currency string) bool {
	acc, err := server.store.GetAccountById(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if acc.Currency != currency {
		err = fmt.Errorf("account [%d] currency mismatch. Was %s but %s", acc.ID, acc.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
