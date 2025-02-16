package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/controller"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/repository/postgresrepo"
	"github.com/myacey/avito-shop/internal/service"
	"github.com/stretchr/testify/require"
)

var (
	itemType = "cup"
	item     = db.Item{ItemID: 1, ItemType: itemType, ItemPrice: 10}

	mockDBUser = db.User{UserID: 1, Username: "mockuser", Password: "mockpassword", Coins: 1000}
)

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	mock.ExpectBegin()  // we're using dbConn.BeginTx in service.BuyItem
	mock.ExpectCommit() // we're using tx.Commit() in service.BuyItem

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := postgresrepo.NewPostgresUserRepo(mockStore)
	mockStore.EXPECT().
		GetUserForUpdate(gomock.Any(), mockDBUser.Username).
		Return(mockDBUser, nil)
	mockStore.EXPECT().
		UpdateUserBalance(gomock.Any(), db.UpdateUserBalanceParams{
			UserID: mockDBUser.UserID,
			Coins:  mockDBUser.Coins - int32(item.ItemPrice),
		}).
		Return(db.User{}, nil)

	inventoryRepo := postgresrepo.NewPostgresInventoryRepo(mockStore)
	mockStore.EXPECT().
		BuyItem(gomock.Any(), db.BuyItemParams{UserID: mockDBUser.UserID, ItemType: itemType}).
		Return(nil)

	transferRepo := postgresrepo.NewPostgresTransferRepo(mockStore)

	storeRepo := postgresrepo.NewPostgresStoreRepo(mockStore)
	mockStore.EXPECT().
		GetItemFromStore(gomock.Any(), itemType).
		Return(item, nil)

	srv := service.NewService(dbConn, userRepo, transferRepo, inventoryRepo, storeRepo, nil, nil, nil)

	handler := controller.NewController(srv)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", mockDBUser.Username)
	c.Params = append(c.Params, gin.Param{Key: "item", Value: itemType})

	req, err := http.NewRequest("GET", "/api/buy/"+itemType, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.BuyItem(c)

	require.Equal(t, http.StatusOK, w.Code)
}
