package postgresrepo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/repository"
	"github.com/stretchr/testify/require"
)

var (
	mockUser1 = db.User{UserID: 1, Username: "mockuser1", Password: "mockpassword", Coins: 1000}
	mockUser2 = db.User{UserID: 2, Username: "mockuser2", Password: "mockpassword", Coins: 1000}
	ErrMock   = errors.New("mock error")
)

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := NewPostgresUserRepo(mockStore)
	testCases := []struct {
		name          string
		username      string
		password      string
		mockBehavior  func(username, password string)
		expectedUser  *db.User
		expectedError error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				mockStore.EXPECT().
					CreateUser(gomock.Any(), gomock.Eq(db.CreateUserParams{Username: username, Password: password})).
					Return(mockUser1, nil)
			},
			expectedUser:  &mockUser1,
			expectedError: nil,
		},
		{
			name:     "Not Unique Username",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username string, password string) {
				mockStore.EXPECT().
					CreateUser(gomock.Any(), gomock.Eq(db.CreateUserParams{Username: username, Password: password})).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			expectedUser:  nil,
			expectedError: repository.ErrUserAlreadyExists,
		},
		{
			name:     "Unknown Error",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				mockStore.EXPECT().
					CreateUser(gomock.Any(), gomock.Eq(db.CreateUserParams{Username: username, Password: password})).
					Return(db.User{}, ErrMock)
			},
			expectedUser:  nil,
			expectedError: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username, tc.password)
			usr, err := userRepo.CreateUser(context.Background(), tc.username, tc.password)

			require.Equal(t, tc.expectedUser, usr)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := NewPostgresUserRepo(mockStore)
	testCases := []struct {
		name          string
		username      string
		mockBehavior  func(username string)
		expectedUser  *db.User
		expectedError error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUser(gomock.Any(), username).
					Return(mockUser1, nil)
			},
			expectedUser:  &mockUser1,
			expectedError: nil,
		},
		{
			name:     "No User",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUser(gomock.Any(), username).
					Return(db.User{}, sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: repository.ErrUserNotFound,
		},
		{
			name:     "Unknown Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUser(gomock.Any(), username).
					Return(db.User{}, ErrMock)
			},
			expectedUser:  nil,
			expectedError: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username)
			usr, err := userRepo.GetUser(context.Background(), tc.username)

			require.Equal(t, tc.expectedUser, usr)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGetUserForUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := NewPostgresUserRepo(mockStore)
	testCases := []struct {
		name          string
		username      string
		mockBehavior  func(username string)
		expectedUser  *db.User
		expectedError error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(mockUser1, nil)
			},
			expectedUser:  &mockUser1,
			expectedError: nil,
		},
		{
			name:     "No User",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(db.User{}, sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: repository.ErrUserNotFound,
		},
		{
			name:     "Unknown Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(db.User{}, ErrMock)
			},
			expectedUser:  nil,
			expectedError: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username)
			usr, err := userRepo.GetUserForUpdate(context.Background(), tc.username)

			require.Equal(t, tc.expectedUser, usr)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUpdateBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := NewPostgresUserRepo(mockStore)
	testCases := []struct {
		name          string
		userID        int32
		newCoinsCount int32
		mockBehavior  func(userID, coinsCount int32)
		expectedUser  *db.User
		expectedError error
	}{
		{
			name:          "OK",
			userID:        1,
			newCoinsCount: 1100,
			mockBehavior: func(userID, newCoinsCount int32) {
				mockStore.EXPECT().
					UpdateUserBalance(gomock.Any(), gomock.Eq(db.UpdateUserBalanceParams{userID, newCoinsCount})).
					Return(db.User{mockUser1.UserID, mockUser1.Username, mockUser1.Password, newCoinsCount}, nil)
			},
			expectedUser:  &db.User{mockUser1.UserID, mockUser1.Username, mockUser1.Password, 1100},
			expectedError: nil,
		},
		{
			name:          "No User",
			userID:        1,
			newCoinsCount: 1100,
			mockBehavior: func(userID, newCoinsCount int32) {
				mockStore.EXPECT().
					UpdateUserBalance(gomock.Any(), gomock.Eq(db.UpdateUserBalanceParams{userID, newCoinsCount})).
					Return(db.User{}, sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: repository.ErrUserNotFound,
		},
		{
			name:          "Unexpected Error",
			userID:        1,
			newCoinsCount: 1100,
			mockBehavior: func(userID, newCoinsCount int32) {
				mockStore.EXPECT().
					UpdateUserBalance(gomock.Any(), gomock.Eq(db.UpdateUserBalanceParams{userID, newCoinsCount})).
					Return(db.User{}, ErrMock)
			},
			expectedUser:  nil,
			expectedError: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.userID, tc.newCoinsCount)
			usr, err := userRepo.UpdateBalance(context.Background(), tc.userID, tc.newCoinsCount)

			require.Equal(t, tc.expectedUser, usr)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUpdateTwoUsersBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := NewPostgresUserRepo(mockStore)
	testCases := []struct {
		name          string
		fromUsername  string
		toUsername    string
		amount        int32
		mockBehavior  func(fromUsername, toUsername string, coins int32)
		expectedUsers []*db.User
		expectedError error
	}{
		{
			name:         "OK",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       100,
			mockBehavior: func(fromUsername, toUsername string, coins int32) {
				mockStore.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), gomock.Eq(db.UpdateTwoUsersBalanceParams{Coins: coins, FromUsername: fromUsername, ToUsername: toUsername})).
					Return([]db.User{
						{mockUser1.UserID, mockUser1.Username, mockUser1.Password, mockUser1.Coins - coins},
						{mockUser2.UserID, mockUser2.Username, mockUser2.Password, mockUser2.Coins + coins},
					}, nil)
			},
			expectedUsers: []*db.User{
				{mockUser1.UserID, mockUser1.Username, mockUser1.Password, mockUser1.Coins - 100},
				{mockUser2.UserID, mockUser2.Username, mockUser2.Password, mockUser2.Coins + 100},
			},
			expectedError: nil,
		},
		{
			name:         "No User",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       100,
			mockBehavior: func(fromUsername, toUsername string, coins int32) {
				mockStore.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), gomock.Eq(db.UpdateTwoUsersBalanceParams{Coins: coins, FromUsername: fromUsername, ToUsername: toUsername})).
					Return([]db.User{}, sql.ErrNoRows)
			},
			expectedUsers: nil,
			expectedError: repository.ErrUserNotFound,
		},
		{
			name:         "Unknown Error",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       100,
			mockBehavior: func(fromUsername, toUsername string, coins int32) {
				mockStore.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), gomock.Eq(db.UpdateTwoUsersBalanceParams{Coins: coins, FromUsername: fromUsername, ToUsername: toUsername})).
					Return([]db.User{}, ErrMock)
			},
			expectedUsers: nil,
			expectedError: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.fromUsername, tc.toUsername, tc.amount)
			usrs, err := userRepo.UpdateTwoUsersBalance(context.Background(), tc.fromUsername, tc.toUsername, tc.amount)

			require.Equal(t, tc.expectedUsers, usrs)
			require.Equal(t, tc.expectedError, err)
		})
	}
}
