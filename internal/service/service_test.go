package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/apperror"
	"github.com/myacey/avito-shop/internal/hasher"
	"github.com/myacey/avito-shop/internal/jwttoken"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/models"
	"github.com/myacey/avito-shop/internal/repository"
	"github.com/stretchr/testify/require"
)

var (
	// USERS
	mockUser1 = db.User{UserID: 1, Username: "mockuser1", Password: "mockpassword", Coins: 1000}
	mockUser2 = db.User{UserID: 2, Username: "mockuser2", Password: "mockpassword", Coins: 1000}

	mockItem = &db.Item{ItemID: 1, ItemType: "mockitem", ItemPrice: 10}
	// Inventory
	mockInventory1  = &db.Inventory{InventoryID: 1, UserID: mockUser1.UserID, ItemType: "mockInventory1", Quantity: 10}
	mockInventory2  = &db.Inventory{InventoryID: 2, UserID: mockUser1.UserID, ItemType: "mockInventory2", Quantity: 10}
	mockInventories = []*db.Inventory{mockInventory1, mockInventory2}

	// Transfers
	tx1 = &db.Transfer{TransferID: 1, FromUsername: mockUser1.Username, ToUsername: mockUser2.Username, Amount: 10}
	tx2 = &db.Transfer{TransferID: 2, FromUsername: mockUser2.Username, ToUsername: mockUser1.Username, Amount: 20}

	ErrMock = errors.New("mock error")
)

func TestAuthorizeUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	transferRepo := mocks.NewMockTransferRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	inventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	storeRepo := mocks.NewMockStoreRepository(ctrl)

	jwtToken := mocks.NewMockTokenMakerInterface(ctrl)

	hashGen := mocks.NewMockHasher(ctrl)

	var dummyDB *sql.DB = nil

	srv := NewService(dummyDB, userRepo, transferRepo, inventoryRepo, storeRepo, sessionRepo, jwtToken, hashGen)

	testCases := []struct {
		name         string
		username     string
		password     string
		mockBehavior func(username, password string)
		expToken     string
		expErr       error
	}{
		{
			name:     "OK Login",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				hashGen.EXPECT().
					Compare(mockUser1.Password, mockUser1.Password).
					Return(nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("valid", nil)
				sessionRepo.EXPECT().
					CreateToken(gomock.Any(), username, "valid", 24*time.Hour).
					Return(nil)
			},
			expToken: "valid",
			expErr:   nil,
		},
		{
			name:     "OK Create New User",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return(mockUser1.Password, nil)
				userRepo.EXPECT().
					CreateUser(gomock.Any(), username, password).
					Return(&mockUser1, nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("valid", nil)
				sessionRepo.EXPECT().
					CreateToken(gomock.Any(), username, "valid", 24*time.Hour).
					Return(nil)
			},
			expToken: "valid",
			expErr:   nil,
		},
		{
			name:     "Err Create User Unknown Error",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return(mockUser1.Password, nil)
				userRepo.EXPECT().
					CreateUser(gomock.Any(), username, password).
					Return(nil, ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to create user", ErrMock),
		},
		{
			name:     "Err Create User Password Too Long",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return("", hasher.ErrToLong)
			},
			expToken: "",
			expErr:   apperror.NewBadReq("password too long", hasher.ErrToLong),
		},
		{
			name:     "Unknwon Err Create User Password",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return("", ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to generate password hash", ErrMock),
		},
		{
			name:     "Err Create Token",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return(mockUser1.Password, nil)
				userRepo.EXPECT().
					CreateUser(gomock.Any(), username, password).
					Return(&mockUser1, nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("", ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to create token for user", ErrMock),
		},
		{
			name:     "Err Session Create Token",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
				hashGen.EXPECT().
					Generate(mockUser1.Password).
					Return(mockUser1.Password, nil)
				userRepo.EXPECT().
					CreateUser(gomock.Any(), username, password).
					Return(&mockUser1, nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("valid", nil)
				sessionRepo.EXPECT().
					CreateToken(gomock.Any(), username, "valid", 24*time.Hour).
					Return(ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to save token", ErrMock),
		},
		{
			name:     "GetUser Unkown Err",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to get user", ErrMock),
		},
		{
			name:     "Craete New Token Err",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				hashGen.EXPECT().
					Compare(mockUser1.Password, mockUser1.Password).
					Return(nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("", ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to create token", ErrMock),
		},
		{
			name:     "Err Save Token",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				hashGen.EXPECT().
					Compare(mockUser1.Password, mockUser1.Password).
					Return(nil)
				jwtToken.EXPECT().
					CreateToken(username).
					Return("valid", nil)
				sessionRepo.EXPECT().
					CreateToken(gomock.Any(), username, "valid", 24*time.Hour).
					Return(ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to save token", ErrMock),
		},
		{
			name:     "Err Compare Password Found User",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				hashGen.EXPECT().
					Compare(mockUser1.Password, mockUser1.Password).
					Return(hasher.ErrDontCompare)
			},
			expToken: "",
			expErr:   apperror.NewNotFound("user not found", hasher.ErrDontCompare),
		},
		{
			name:     "Err Unknown Compare Password Found User",
			username: mockUser1.Username,
			password: mockUser1.Password,
			mockBehavior: func(username, password string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				hashGen.EXPECT().
					Compare(mockUser1.Password, mockUser1.Password).
					Return(ErrMock)
			},
			expToken: "",
			expErr:   apperror.NewInternal("failed to compare passwords", ErrMock),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username, tc.password)

			tok, err := srv.AuthorizeUser(context.Background(), tc.username, tc.password)

			require.Equal(t, tc.expToken, tok)
			require.Equal(t, tc.expErr, err)
		})
	}
}

func TestCheckAuthToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	transferRepo := mocks.NewMockTransferRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	inventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	storeRepo := mocks.NewMockStoreRepository(ctrl)

	jwtToken := mocks.NewMockTokenMakerInterface(ctrl)

	var dummyDB *sql.DB = nil

	srv := NewService(dummyDB, userRepo, transferRepo, inventoryRepo, storeRepo, sessionRepo, jwtToken, nil)

	testCases := []struct {
		name         string
		token        string
		mockBehavior func(token string)
		expUsername  string
		expErr       error
	}{
		{
			name:  "OK",
			token: "valid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("valid").
					Return(mockUser1.Username, nil)
				sessionRepo.EXPECT().
					GetToken(gomock.Any(), mockUser1.Username).
					Return("valid", nil)
			},
			expUsername: mockUser1.Username,
			expErr:      nil,
		},
		{
			name:  "Err Invalid Key",
			token: "invalid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("invalid").
					Return("", jwttoken.ErrInvalidKey)
			},
			expUsername: "",
			expErr:      apperror.NewUnauthorized(jwttoken.ErrInvalidKey.Error(), nil),
		},
		{
			name:  "Err Expired Key",
			token: "expired",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("expired").
					Return("", jwttoken.ErrTokenExpired)
			},
			expUsername: "",
			expErr:      apperror.NewUnauthorized(jwttoken.ErrTokenExpired.Error(), nil),
		},
		{
			name:  "Unknown Token Err",
			token: "valid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("valid").
					Return("", ErrMock)
			},
			expUsername: "",
			expErr:      apperror.NewInternal("failed to verify token", ErrMock),
		},
		{
			name:  "Err Token Not Found",
			token: "valid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("valid").
					Return(mockUser1.Username, nil)
				sessionRepo.EXPECT().
					GetToken(gomock.Any(), mockUser1.Username).
					Return("", repository.ErrTokenNotFound)
			},
			expUsername: "",
			expErr:      apperror.NewUnauthorized("invalid token", repository.ErrTokenNotFound),
		},
		{
			name:  "Err Session Unknown Error",
			token: "valid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("valid").
					Return(mockUser1.Username, nil)
				sessionRepo.EXPECT().
					GetToken(gomock.Any(), mockUser1.Username).
					Return("", ErrMock)
			},
			expUsername: "",
			expErr:      apperror.NewInternal("failed to find token", ErrMock),
		},
		{
			name:  "Invalid Token",
			token: "invalid",
			mockBehavior: func(token string) {
				jwtToken.EXPECT().
					VerifyToken("invalid").
					Return(mockUser1.Username, nil)
				sessionRepo.EXPECT().
					GetToken(gomock.Any(), mockUser1.Username).
					Return("valid", nil)
			},
			expUsername: "",
			expErr:      apperror.NewUnauthorized("invalid token", ErrInvalidToken),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.token)

			usrname, err := srv.CheckAuthToken(context.Background(), tc.token)

			require.Equal(t, tc.expUsername, usrname)
			require.Equal(t, tc.expErr, err)
		})
	}
}

func TestGetFullUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	transferRepo := mocks.NewMockTransferRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	inventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	storeRepo := mocks.NewMockStoreRepository(ctrl)

	jwtToken := mocks.NewMockTokenMakerInterface(ctrl)

	var dummyDB *sql.DB = nil

	srv := NewService(dummyDB, userRepo, transferRepo, inventoryRepo, storeRepo, sessionRepo, jwtToken, nil)

	testCases := []struct {
		name         string
		username     string
		mockBehavior func(username string)
		expUser      *models.User
		expErr       error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				inventoryRepo.EXPECT().
					GetInventory(gomock.Any(), mockUser1.UserID).
					Return(mockInventories, nil)
				transferRepo.EXPECT().
					GetTransfersWithUser(gomock.Any(), mockUser1.Username).
					Return([]*db.Transfer{
						{
							TransferID:   tx1.TransferID,
							FromUsername: mockUser1.Username,
							ToUsername:   mockUser2.Username,
							Amount:       tx1.Amount,
						},
						{
							TransferID:   tx2.TransferID,
							FromUsername: mockUser2.Username,
							ToUsername:   mockUser1.Username,
							Amount:       tx2.Amount,
						},
					}, nil)
			},
			expUser: &models.User{
				ID:       mockUser1.UserID,
				Username: mockUser1.Username,
				Coins:    mockUser1.Coins,
				Inventory: []*models.InventoryItem{
					{mockInventory1.ItemType, mockInventory1.Quantity},
					{mockInventory2.ItemType, mockInventory2.Quantity},
				},
				EntryHistory: map[string]interface{}{
					"sent":     []*OutcomeEntry{{mockUser2.Username, tx1.Amount}},
					"received": []*IncomeEntry{{mockUser2.Username, tx2.Amount}},
				},
			},
			expErr: nil,
		},
		{
			name:     "Err User Not Found",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
			},
			expUser: nil,
			expErr:  apperror.NewNotFound("user not found", repository.ErrUserNotFound),
		},
		{
			name:     "Unknown User Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(nil, ErrMock)
			},
			expUser: nil,
			expErr:  apperror.NewInternal("failed to get user", ErrMock),
		},
		{
			name:     "Unknown Inventory Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				inventoryRepo.EXPECT().
					GetInventory(gomock.Any(), mockUser1.UserID).
					Return(nil, ErrMock)
			},
			expUser: nil,
			expErr:  apperror.NewInternal("failed to get user inventory", ErrMock),
		},
		{
			name:     "Unknown Transfer Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				userRepo.EXPECT().
					GetUser(gomock.Any(), username).
					Return(&mockUser1, nil)
				inventoryRepo.EXPECT().
					GetInventory(gomock.Any(), mockUser1.UserID).
					Return(mockInventories, nil)
				transferRepo.EXPECT().
					GetTransfersWithUser(gomock.Any(), mockUser1.Username).
					Return(nil, ErrMock)
			},
			expUser: nil,
			expErr:  apperror.NewInternal("failed to find transactions", ErrMock),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username)

			usr, err := srv.GetFullUserInfo(context.Background(), tc.username)

			require.Equal(t, tc.expUser, usr)
			require.Equal(t, tc.expErr, err)
		})
	}
}

func TestSendCoin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	transferRepo := mocks.NewMockTransferRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	inventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	storeRepo := mocks.NewMockStoreRepository(ctrl)

	jwtToken := mocks.NewMockTokenMakerInterface(ctrl)

	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	srv := NewService(dbConn, userRepo, transferRepo, inventoryRepo, storeRepo, sessionRepo, jwtToken, nil)

	testCases := []struct {
		name         string
		fromUsername string
		toUsername   string
		amount       int32
		mockBehavior func(fromUsername, toUsername string, amount int32)
		expErr       error
	}{
		{
			name:         "OK",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       tx1.Amount,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				userRepo.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), fromUsername, toUsername, amount).
					Return([]*db.User{}, nil)
				transferRepo.EXPECT().
					CreateMoneyTransfer(gomock.Any(), fromUsername, toUsername, amount).
					Return(nil, nil)
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			expErr: nil,
		},
		{
			name:         "Invalid Amount",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       -100,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {},
			expErr:       apperror.NewBadReq("send coins amont must be positive", nil),
		},
		{
			name:         "Err Transfer",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       tx1.Amount,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				mock.ExpectBegin().WillReturnError(ErrMock)
			},
			expErr: apperror.NewInternal("failed to send coins", ErrMock),
		},
		{
			name:         "Err User Not Found",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       tx1.Amount,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				userRepo.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), fromUsername, toUsername, amount).
					Return([]*db.User{}, repository.ErrUserNotFound)
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			expErr: apperror.NewNotFound(fmt.Sprintf("users not found: %s, %s", mockUser1.Username, mockUser2.Username), repository.ErrUserNotFound),
		},
		{
			name:         "Err Unknown",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       tx1.Amount,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				userRepo.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), fromUsername, toUsername, amount).
					Return([]*db.User{}, ErrMock)
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			expErr: apperror.NewInternal("failed to make money transaction", ErrMock),
		},
		{
			name:         "Err Create Money Transfer",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       tx1.Amount,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				userRepo.EXPECT().
					UpdateTwoUsersBalance(gomock.Any(), fromUsername, toUsername, amount).
					Return([]*db.User{}, nil)
				transferRepo.EXPECT().
					CreateMoneyTransfer(gomock.Any(), fromUsername, toUsername, amount).
					Return(nil, ErrMock)
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			expErr: apperror.NewInternal("failed to create transfer", ErrMock),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.fromUsername, tc.toUsername, tc.amount)

			err := srv.SendCoin(context.Background(), tc.fromUsername, tc.toUsername, tc.amount)

			require.Equal(t, tc.expErr, err)
		})
	}
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	transferRepo := mocks.NewMockTransferRepository(ctrl)
	sessionRepo := mocks.NewMockSessionRepository(ctrl)
	inventoryRepo := mocks.NewMockInventoryRepository(ctrl)
	storeRepo := mocks.NewMockStoreRepository(ctrl)

	jwtToken := mocks.NewMockTokenMakerInterface(ctrl)

	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	srv := NewService(dbConn, userRepo, transferRepo, inventoryRepo, storeRepo, sessionRepo, jwtToken, nil)

	testCases := []struct {
		name         string
		username     string
		itemName     string
		mockBehavior func(username, itemName string)
		expErr       error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectCommit()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(&mockUser1, nil)
				userRepo.EXPECT().
					UpdateBalance(gomock.Any(), mockUser1.UserID, mockUser1.Coins-int32(mockItem.ItemPrice)).
					Return(nil, nil)
				inventoryRepo.EXPECT().
					AddItemToInventory(gomock.Any(), mockUser1.UserID, itemName).
					Return(nil)
			},
			expErr: nil,
		},
		{
			name:     "Err Invalid Item Name",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(nil, repository.ErrInvalidItemName)
			},
			expErr: apperror.NewBadReq("invalid item name", repository.ErrInvalidItemName),
		},
		{
			name:     "Unknown Err Get Item Info",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(nil, ErrMock)
			},
			expErr: apperror.NewInternal("failed to get item info", ErrMock),
		},
		{
			name:     "Err TX",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin().WillReturnError(ErrMock)
			},
			expErr: apperror.NewInternal("failed to buy item", ErrMock),
		},
		{
			name:     "Err User Not Found",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectRollback()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(nil, repository.ErrUserNotFound)
			},
			expErr: apperror.NewNotFound("user not found", repository.ErrUserNotFound),
		},
		{
			name:     "Err User Not Found",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectRollback()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(nil, ErrMock)
			},
			expErr: apperror.NewInternal("failed to get user", ErrMock),
		},
		{
			name:     "Not Enough Coins",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectRollback()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(&db.User{Coins: 0}, nil)
			},
			expErr: apperror.NewBadReq("not enough money", ErrNotEnoughMoney),
		},
		{
			name:     "Err Update",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectRollback()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(&mockUser1, nil)
				userRepo.EXPECT().
					UpdateBalance(gomock.Any(), mockUser1.UserID, mockUser1.Coins-int32(mockItem.ItemPrice)).
					Return(nil, ErrMock)
			},
			expErr: apperror.NewInternal("failed to update balance", ErrMock),
		},
		{
			name:     "Err Add Item",
			username: mockUser1.Username,
			mockBehavior: func(username, itemName string) {
				storeRepo.EXPECT().
					GetItemInfo(gomock.Any(), itemName).
					Return(mockItem, nil)
				mock.ExpectBegin()
				mock.ExpectRollback()
				userRepo.EXPECT().
					GetUserForUpdate(gomock.Any(), username).
					Return(&mockUser1, nil)
				userRepo.EXPECT().
					UpdateBalance(gomock.Any(), mockUser1.UserID, mockUser1.Coins-int32(mockItem.ItemPrice)).
					Return(nil, nil)
				inventoryRepo.EXPECT().
					AddItemToInventory(gomock.Any(), mockUser1.UserID, itemName).
					Return(ErrMock)
			},
			expErr: apperror.NewInternal("failed to add item to inventory", ErrMock),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username, tc.itemName)

			err := srv.BuyItem(context.Background(), tc.username, tc.itemName)

			require.Equal(t, tc.expErr, err)
		})
	}
}
