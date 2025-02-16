package postgresrepo

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/repository"
	"github.com/stretchr/testify/require"
)

var (
	mockTransfer1 = db.Transfer{TransferID: 1, FromUsername: "mockuser1", ToUsername: "mockuser2", Amount: 10}
	mockTransfer2 = db.Transfer{TransferID: 2, FromUsername: "mockuser2", ToUsername: "mockuser1", Amount: 10}
)

func TestCreateMoneyTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	transferRepo := NewPostgresTransferRepo(mockStore)

	testCases := []struct {
		name         string
		fromUsername string
		toUsername   string
		amount       int32
		mockBehavior func(fromUsername, toUsername string, amount int32)
		expAns       *db.Transfer
		expErr       error
	}{
		{
			name:         "OK",
			fromUsername: mockUser1.Username,
			toUsername:   mockUser2.Username,
			amount:       10,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				mockStore.EXPECT().
					CreateMoneyTransfer(gomock.Any(), gomock.Eq(db.CreateMoneyTransferParams{FromUsername: fromUsername, ToUsername: toUsername, Amount: amount})).
					Return(mockTransfer1, nil)
			},
			expAns: &mockTransfer1,
			expErr: nil,
		},
		{
			name:         "Error",
			toUsername:   mockUser1.Username,
			fromUsername: mockUser2.Username,
			amount:       10,
			mockBehavior: func(fromUsername, toUsername string, amount int32) {
				mockStore.EXPECT().
					CreateMoneyTransfer(gomock.Any(), gomock.Eq(db.CreateMoneyTransferParams{FromUsername: fromUsername, ToUsername: toUsername, Amount: amount})).
					Return(db.Transfer{}, ErrMock)
			},
			expAns: nil,
			expErr: ErrMock,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			ts.mockBehavior(ts.fromUsername, ts.toUsername, ts.amount)

			tx, err := transferRepo.CreateMoneyTransfer(context.Background(), ts.fromUsername, ts.toUsername, ts.amount)

			require.Equal(t, tx, ts.expAns)
			require.Equal(t, err, ts.expErr)
		})
	}
}

func TestGetTransfersWithUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	transferRepo := NewPostgresTransferRepo(mockStore)

	testCases := []struct {
		name         string
		username     string
		mockBehavior func(username string)
		expAns       []*db.Transfer
		expErr       error
	}{
		{
			name:     "OK",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetTransfersWithUser(gomock.Any(), username).
					Return([]db.Transfer{
						{
							mockTransfer1.TransferID,
							mockTransfer1.FromUsername,
							mockTransfer1.ToUsername,
							mockTransfer1.Amount,
						},
						{
							mockTransfer2.TransferID,
							mockTransfer2.FromUsername,
							mockTransfer2.ToUsername,
							mockTransfer2.Amount,
						},
					}, nil)
			},
			expAns: []*db.Transfer{
				{
					mockTransfer1.TransferID,
					mockTransfer1.FromUsername,
					mockTransfer1.ToUsername,
					mockTransfer1.Amount,
				},
				{
					mockTransfer2.TransferID,
					mockTransfer2.FromUsername,
					mockTransfer2.ToUsername,
					mockTransfer2.Amount,
				},
			},
			expErr: nil,
		},
		{
			name:     "Err No Transfers",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetTransfersWithUser(gomock.Any(), username).
					Return([]db.Transfer{}, sql.ErrNoRows)
			},
			expAns: nil,
			expErr: repository.ErrNoTransfers,
		},
		{
			name:     "Unexpected Error",
			username: mockUser1.Username,
			mockBehavior: func(username string) {
				mockStore.EXPECT().
					GetTransfersWithUser(gomock.Any(), username).
					Return([]db.Transfer{}, ErrMock)
			},
			expAns: nil,
			expErr: ErrMock,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			ts.mockBehavior(ts.username)

			tx, err := transferRepo.GetTransfersWithUser(context.Background(), mockUser1.Username)

			require.Equal(t, tx, ts.expAns)
			require.Equal(t, err, ts.expErr)
		})
	}
}
