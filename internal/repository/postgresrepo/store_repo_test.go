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

var mockItem = db.Item{ItemID: 1, ItemType: "mockType", ItemPrice: 10}

func TestGetItemInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	storeRepo := NewPostgresStoreRepo(mockStore)

	testCases := []struct {
		name         string
		itemName     string
		mockBehavior func(itemName string)
		expAns       *db.Item
		expErr       error
	}{
		{
			name:     "OK",
			itemName: mockItem.ItemType,
			mockBehavior: func(itemName string) {
				mockStore.EXPECT().
					GetItemFromStore(gomock.Any(), itemName).
					Return(mockItem, nil)
			},
			expAns: &mockItem,
			expErr: nil,
		},
		{
			name:     "Invalid Item Name",
			itemName: "invalid",
			mockBehavior: func(itemName string) {
				mockStore.EXPECT().
					GetItemFromStore(gomock.Any(), itemName).
					Return(db.Item{}, sql.ErrNoRows)
			},
			expAns: nil,
			expErr: repository.ErrInvalidItemName,
		},
		{
			name:     "Unknown Error",
			itemName: "invalid",
			mockBehavior: func(itemName string) {
				mockStore.EXPECT().
					GetItemFromStore(gomock.Any(), itemName).
					Return(db.Item{}, ErrMock)
			},
			expAns: nil,
			expErr: ErrMock,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			ts.mockBehavior(ts.itemName)

			item, err := storeRepo.GetItemInfo(context.Background(), ts.itemName)

			require.Equal(t, item, ts.expAns)
			require.Equal(t, err, ts.expErr)
		})
	}
}
