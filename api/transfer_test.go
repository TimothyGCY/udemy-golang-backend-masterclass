package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	mockdb "learn.bleckshiba/banking/db/mock"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/enum"
	"net/http"
	"net/http/httptest"
	"testing"
)

type transferTestCase struct {
	Name        string
	Body        gin.H
	BuildStub   func(store *mockdb.MockStore)
	PostRequest func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestNewTransfer(t *testing.T) {
	acc1, acc2, acc3 := randomAccount(), randomAccount(), randomAccount()
	acc1.Currency = enum.MYR
	acc2.Currency = enum.MYR
	acc3.Currency = enum.USD

	rq := gin.H{
		"fromAccount": acc1.ID,
		"toAccount":   acc2.ID,
		"amount":      10.0,
		"currency":    enum.MYR,
	}

	testCases := []transferTestCase{
		{
			Name: "OK",
			Body: rq,
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), db.TransferTxParam{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        10.0,
					}).Times(1)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			Name: "NotFound/FromAccountNotFound",
			Body: rq,
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			Name: "NotFound/ToAccountNotFound",
			Body: rq,
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			Name: "BadRequest/FromAccountCurrencyMismatch",
			Body: gin.H{
				"fromAccount": acc1.ID,
				"toAccount":   acc2.ID,
				"amount":      10.0,
				"currency":    enum.USD,
			},
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{

			Name: "BadRequest/ToAccountCurrencyMismatch",
			Body: gin.H{
				"fromAccount": acc1.ID,
				"toAccount":   acc3.ID,
				"amount":      10.0,
				"currency":    enum.MYR,
			},
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc3.ID)).
					Times(1).Return(acc3, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			Name: "BadRequest/InsufficientBalance",
			Body: gin.H{
				"fromAccount": acc1.ID,
				"toAccount":   acc2.ID,
				"amount":      float64(1000000),
				"currency":    enum.MYR,
			},
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			Name: "InternalServerError/FailedToGetAccount",
			Body: rq,
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			Name: "InternalServerError/TransactionFailed",
			Body: rq,
			BuildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).Return(acc1, nil)
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), db.TransferTxParam{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        10.0,
					}).Times(1).Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			PostRequest: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	server, err := NewServer(store)
	assert.NoError(t, err)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.Name, func(t *testing.T) {
			tc.BuildStub(store)
			recorder := httptest.NewRecorder()

			marshal, err := json.Marshal(tc.Body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(marshal))
			assert.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.PostRequest(t, recorder)
		})
	}
}
