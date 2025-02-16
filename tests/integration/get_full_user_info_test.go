package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/controller"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/models"
	"github.com/myacey/avito-shop/internal/repository/postgresrepo"
	"github.com/myacey/avito-shop/internal/service"
	"github.com/stretchr/testify/require"
)

var mockInventories = []db.Inventory{
	{InventoryID: 1, UserID: mockDBUser.UserID, ItemType: "mockItem1", Quantity: 10},
	{InventoryID: 2, UserID: mockDBUser.UserID, ItemType: "mockItem2", Quantity: 20},
}

var mockTransfers = []db.Transfer{
	{TransferID: 1, FromUsername: "mockuser1", ToUsername: "mockuser2", Amount: 100},
	{TransferID: 2, FromUsername: "mockuser2", ToUsername: "mockuser1", Amount: 200},
}

func TestGetFullUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockQuerier(ctrl)
	userRepo := postgresrepo.NewPostgresUserRepo(mockStore)
	mockStore.EXPECT().
		GetUser(gomock.Any(), "mockuser1").
		Return(mockDBUser, nil)

	inventoryRepo := postgresrepo.NewPostgresInventoryRepo(mockStore)
	mockStore.EXPECT().
		GetInventory(gomock.Any(), mockDBUser.UserID).
		Return(mockInventories, nil)

	transferRepo := postgresrepo.NewPostgresTransferRepo(mockStore)
	mockStore.EXPECT().
		GetTransfersWithUser(gomock.Any(), "mockuser1").
		Return(mockTransfers, nil)

	srv := service.NewService(nil, userRepo, transferRepo, inventoryRepo, nil, nil, nil, nil)

	handler := controller.NewController(srv)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("username", "mockuser1")

	req, err := http.NewRequest("GET", "/api/info/", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.GetFullUserInfo(c)

	require.Equal(t, http.StatusOK, w.Code)

	expUser := &models.User{
		ID:       1,
		Username: "mockuser1",
		Coins:    mockDBUser.Coins,
		Inventory: []*models.InventoryItem{
			{"mockItem1", 10},
			{"mockItem2", 20},
		},
		EntryHistory: map[string]interface{}{
			"received": []*service.IncomeEntry{
				{"mockuser2", 200},
			},
			"sent": []*service.OutcomeEntry{
				{"mockuser2", 100},
			},
		},
		Password: "",
	}
	expectedJSON, err := json.Marshal(expUser)
	require.NoError(t, err)
	actualJSON := w.Body.Bytes()

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}
