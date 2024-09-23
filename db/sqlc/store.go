package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
  Querier
	TransferTx(context.Context, TransferTxParam) (TransferTxResult, error)
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{db: db, Queries: New(db)}
}

func (s *SQLStore) execTrx(ctx context.Context, fn func(queries *Queries) error) error {
	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(trx)
	err = fn(q)
	if err != nil {
		if rollbackError := trx.Rollback(); rollbackError != nil {
			return fmt.Errorf("transaction error: %v\n", err.Error())
		}
		return err
	}
	return trx.Commit()
}

type TransferTxParam struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"fromAccount"`
	ToAccount   Account  `json:"toAccount"`
	FromEntry   Entry    `json:"fromEntry"`
	ToEntry     Entry    `json:"toEntry"`
}

func (s *SQLStore) TransferTx(ctx context.Context, arg TransferTxParam) (TransferTxResult, error) {
	var result TransferTxResult

	err := s.execTrx(ctx, func(q *Queries) error {
		var err error
		if result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		}); err != nil {
			return err
		}

		if result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		}); err != nil {
			return err
		}

		if result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		}); err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err =
				transferToAcc(transferToAccParam{
					context:  ctx,
					queries:  q,
					account1: arg.FromAccountID,
					account2: arg.ToAccountID,
					amount1:  arg.Amount,
					amount2:  -arg.Amount,
				})
		} else {
			result.ToAccount, result.FromAccount, err =
				transferToAcc(transferToAccParam{
					context:  ctx,
					queries:  q,
					account1: arg.ToAccountID,
					account2: arg.FromAccountID,
					amount1:  -arg.Amount,
					amount2:  arg.Amount,
				})
		}

		return nil
	})

	return result, err
}

type transferToAccParam struct {
	context  context.Context
	queries  *Queries
	account1 int64
	amount1  float64
	account2 int64
	amount2  float64
}

func transferToAcc(param transferToAccParam) (acc1, acc2 Account, err error) {
	acc1, err = param.queries.TransferMoney(param.context, TransferMoneyParams{
		ID:     param.account1,
		Amount: param.amount1,
	})
	if err != nil {
		return
	}

	acc2, err = param.queries.TransferMoney(param.context, TransferMoneyParams{
		ID:     param.account2,
		Amount: param.amount2,
	})

	return
}
