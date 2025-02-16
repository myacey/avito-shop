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
	mockInventory1 = db.Inventory{InventoryID: 1, UserID: mockUser1.UserID, ItemType: "mockType1", Quantity: 10}
	mockInventory2 = db.Inventory{InventoryID: 2, UserID: mockUser1.UserID, ItemType: "mockType2", Quantity: 20}
)

func TestAddItemToInventory(t *testing.T) {
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockStore := mocks.NewMockQuerier(ctlr)
	inventoryRepo := NewPostgresInventoryRepo(mockStore)

	testCases := []struct {
		name         string
		userID       int32
		itemType     string
		mockBehavior func(userID int32, itemType string)
		expRes       error
	}{
		{
			name:     "OK",
			userID:   mockUser1.UserID,
			itemType: mockItem.ItemType,
			mockBehavior: func(userID int32, itemType string) {
				mockStore.EXPECT().
					BuyItem(gomock.Any(), gomock.Eq(db.BuyItemParams{
						UserID:   mockUser1.UserID,
						ItemType: mockItem.ItemType,
					})).
					Return(nil)
			},
			expRes: nil,
		},
		{
			name:     "Unknown Error",
			userID:   mockUser1.UserID,
			itemType: mockItem.ItemType,
			mockBehavior: func(userID int32, itemType string) {
				mockStore.EXPECT().
					BuyItem(gomock.Any(), gomock.Eq(db.BuyItemParams{
						UserID:   mockUser1.UserID,
						ItemType: mockItem.ItemType,
					})).
					Return(ErrMock)
			},
			expRes: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.userID, tc.itemType)

			res := inventoryRepo.AddItemToInventory(context.Background(), tc.userID, tc.itemType)

			require.Equal(t, res, tc.expRes)
		})
	}
}

func TestGetInventory(t *testing.T) {
	ctlr := gomock.NewController(t)
	defer ctlr.Finish()

	mockStore := mocks.NewMockQuerier(ctlr)
	inventoryRepo := NewPostgresInventoryRepo(mockStore)

	testCases := []struct {
		name         string
		userID       int32
		mockBehavior func(userID int32)
		expRes       []*db.Inventory
		expErr       error
	}{
		{
			name:   "OK",
			userID: mockUser1.UserID,
			mockBehavior: func(userID int32) {
				mockStore.EXPECT().
					GetInventory(gomock.Any(), userID).
					Return([]db.Inventory{mockInventory1, mockInventory2}, nil)
			},
			expRes: []*db.Inventory{&mockInventory1, &mockInventory2},
			expErr: nil,
		},
		{
			name:   "No Inventory",
			userID: mockUser1.UserID,
			mockBehavior: func(userID int32) {
				mockStore.EXPECT().
					GetInventory(gomock.Any(), userID).
					Return([]db.Inventory{}, sql.ErrNoRows)
			},
			expRes: nil,
			expErr: repository.ErrNoInventoryItems,
		},
		{
			name:   "Unexpected Error",
			userID: mockUser1.UserID,
			mockBehavior: func(userID int32) {
				mockStore.EXPECT().
					GetInventory(gomock.Any(), userID).
					Return([]db.Inventory{}, ErrMock)
			},
			expRes: nil,
			expErr: ErrMock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.userID)

			inv, err := inventoryRepo.GetInventory(context.Background(), tc.userID)

			require.Equal(t, inv, tc.expRes)
			require.Equal(t, err, tc.expErr)
		})
	}
}
