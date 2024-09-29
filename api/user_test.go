package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	pg "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	mockdb "learn.bleckshiba/banking/db/mock"
	db "learn.bleckshiba/banking/db/sqlc"
	"learn.bleckshiba/banking/util"
)

type eqCreateUserParamMatcher struct {
	arg      db.CreateUserParams
	password string
}

func EqCreateUserParam(arg db.CreateUserParams, password string) gomock.Matcher {
	return &eqCreateUserParamMatcher{
		arg: arg, password: password}
}

func (m *eqCreateUserParamMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	if err := util.CheckPassword(m.password, arg.HashedPassword); err != nil {
		return false
	}

	m.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(m.arg, arg)
}

func (m *eqCreateUserParamMatcher) String() string {
	return fmt.Sprintf("is equal to %v", m.arg)
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name         string
		body         gin.H
		buildStubs   func(store *mockdb.MockStore)
		postResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					Email:          user.Email,
					HashedPassword: user.HashedPassword,
					FullName:       user.FullName,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParam(arg, password)).
					Times(1).Return(user, nil)
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					Email:          user.Email,
					HashedPassword: user.HashedPassword,
					FullName:       user.FullName,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParam(arg, password)).
					Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Forbidden/DuplicateUserName",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					Email:          user.Email,
					HashedPassword: user.HashedPassword,
					FullName:       user.FullName,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParam(arg, password)).
					Times(1).Return(db.User{}, &pg.Error{Code: "23505"})
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "BadRequest/InvalidUsername",
			body: gin.H{
				"username": "invalid-username", // alphanumeric only
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest/InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    "fake email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequest/PasswordTooShort",
			body: gin.H{
				"username": user.Username,
				"password": "pass",
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			postResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
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

			rqBody, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/user", bytes.NewReader(rqBody))
			assert.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.postResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (db.User, string) {
	g := util.NewGenerator()
	rawPassword := g.RandomPassword()
	password, err := util.HashPassword(rawPassword)
	assert.NoError(t, err)

	return db.User{
		ID:             g.RandomInt64(1, 1000),
		Username:       g.RandomUsername(),
		FullName:       g.RandomName(),
		Email:          g.RandomEmail(),
		HashedPassword: password,
	}, rawPassword
}
