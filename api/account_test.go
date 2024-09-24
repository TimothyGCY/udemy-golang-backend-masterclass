package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	mockdb "learn.bleckshiba/banking/db/mock"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/util"
)

func TestGetAccountByID(t *testing.T) {
	acc := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(acc, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, acc)
			},
		},
		{
			name:      "NotFound",
			accountID: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: acc.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccountById(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	server := NewServer(store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	acc := randomAccount()
	acc.Balance = 0.0

	jsonData := fmt.Sprintf(`{"owner": "%s", "currency": "%s"}`, acc.Owner, acc.Currency)
	fmt.Println(jsonData)
	data := []byte(jsonData)

	param := db.CreateAccountParams{
		Owner:    acc.Owner,
		Currency: acc.Currency,
		Balance:  0.0,
	}

	testCases := []struct {
		name          string
		arg           []byte
		stubs         func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			arg:  data,
			stubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(param)).
					Times(1).
					Return(acc, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, acc)
			},
		},
		{
			name: "InternalServerError",
			arg:  data,
			stubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(param)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			arg:  []byte(`{"abc": "def"}`),
			stubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := NewServer(store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.stubs(store)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(tc.arg))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccounts(t *testing.T) {
	accounts := []db.Account{}
	for i := 0; i < 20; i++ {
		accounts = append(accounts, randomAccount())
	}

	// Define pagination parameters for different test cases
	var pageSize, currentPage int64 = 10, 1

	testCases := []struct {
		name          string
		pageSize      int64
		currentPage   int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			pageSize:    pageSize,
			currentPage: currentPage,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(db.GetAccountsParams{
						Limit:  pageSize,
						Offset: (currentPage - 1) * pageSize,
					})).
					Times(1).
					Return(accounts[:pageSize], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts[:pageSize])
			},
		},
		{
			name:        "NoResults",
			pageSize:    pageSize,
			currentPage: 3, // This would exceed the total number of accounts (only 20 accounts exist)
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(db.GetAccountsParams{
						Limit:  pageSize,
						Offset: (3 - 1) * pageSize, // Offset will exceed the number of available accounts
					})).
					Times(1).
					Return([]db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, []db.Account{})
			},
		},
		{
			name:        "InternalServerError",
			pageSize:    pageSize,
			currentPage: currentPage,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Eq(db.GetAccountsParams{
						Limit:  pageSize,
						Offset: (currentPage - 1) * pageSize,
					})).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "BadRequest_InvalidPageSize",
			pageSize:    0,
			currentPage: 1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "BadRequest_InvalidPageNumber",
			pageSize:    pageSize,
			currentPage: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := NewServer(store)

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)

			// Prepare HTTP request
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts?page=%d&size=%d", tc.currentPage, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	random := util.NewGenerator()
	return db.Account{
		ID:       random.RandomInt64(1, 1000),
		Owner:    random.RandomUsername(),
		Balance:  random.RandomMoney(),
		Currency: random.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var acc db.Account
	err = json.Unmarshal(data, &acc)
	require.NoError(t, err)
	require.Equal(t, account, acc)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var accs []db.Account
	err = json.Unmarshal(data, &accs)
	require.NoError(t, err)
	require.Equal(t, accounts, accs)
}
