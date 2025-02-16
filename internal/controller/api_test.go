package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/myacey/avito-shop/internal/mocks"
	"github.com/myacey/avito-shop/internal/models"
	"github.com/stretchr/testify/require"
)

var (
	ErrMock  = errors.New("mock error")
	mockUser = models.User{1, "mockuser", "", 1000, nil, nil}
)

func TestAuthorize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := mocks.NewMockInterface(ctrl)
	handler := NewController(mockSrv)

	testCases := []struct {
		name         string
		req          authReq
		mockBehavior func(req authReq)
		expStatus    int
		expAns       interface{}
	}{
		{
			name: "OK",
			req:  authReq{"mockuser", "mockpassword"},
			mockBehavior: func(req authReq) {
				mockSrv.EXPECT().
					AuthorizeUser(gomock.Any(), req.Username, req.Password).
					Return("valid", nil)
			},
			expStatus: http.StatusOK,
			expAns:    gin.H{"token": "valid"},
		},
		{
			name: "Err Service",
			req:  authReq{"mockuser", "mockpassword"},
			mockBehavior: func(req authReq) {
				mockSrv.EXPECT().
					AuthorizeUser(gomock.Any(), req.Username, req.Password).
					Return("", ErrMock)
			},
			expStatus: http.StatusInternalServerError,
			expAns:    gin.H{"errors": "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.req)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqMarshalled, err := json.Marshal(tc.req)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "api/auth", bytes.NewBuffer(reqMarshalled))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.Authorize(c)

			require.Equal(t, tc.expStatus, w.Code)
			crResp, err := json.Marshal(tc.expAns)
			require.NoError(t, err)
			require.Equal(t, crResp, w.Body.Bytes())
		})
	}
}

func TestGetFullUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := mocks.NewMockInterface(ctrl)
	handler := NewController(mockSrv)

	testCases := []struct {
		name         string
		username     string
		skipUsername bool
		mockBehavior func(username string)
		expStatus    int
		expAns       interface{}
	}{
		{
			name:     "OK",
			username: "mockuser",
			mockBehavior: func(username string) {
				mockSrv.EXPECT().
					GetFullUserInfo(gomock.Any(), username).
					Return(&mockUser, nil)
			},
			expStatus: http.StatusOK,
			expAns:    mockUser,
		},
		{
			name:         "Err No User",
			username:     "",
			skipUsername: true,
			mockBehavior: func(username string) {
			},
			expStatus: http.StatusInternalServerError,
			expAns:    gin.H{"errors": "no username in token"},
		},
		{
			name:     "Err Service",
			username: "mockuser",
			mockBehavior: func(username string) {
				mockSrv.EXPECT().
					GetFullUserInfo(gomock.Any(), username).
					Return(nil, ErrMock)
			},
			expStatus: http.StatusInternalServerError,
			expAns:    gin.H{"errors": "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if !tc.skipUsername {
				c.Set("username", tc.username)
			}

			req, err := http.NewRequest("GET", "/api/info", nil)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.GetFullUserInfo(c)

			require.Equal(t, tc.expStatus, w.Code)

			if tc.expStatus == http.StatusOK {
				crResp, err := json.Marshal(tc.expAns)
				require.NoError(t, err)
				require.Equal(t, crResp, w.Body.Bytes())
			}
		})
	}
}

func TestSendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := mocks.NewMockInterface(ctrl)
	handler := NewController(mockSrv)

	testCases := []struct {
		name         string
		username     string
		skipUsername bool
		req          sendCoinReq
		mockBehavior func(username string, req sendCoinReq)
		expStatus    int
		expAns       interface{}
	}{
		{
			name:     "OK",
			username: "mockuser",
			mockBehavior: func(username string, req sendCoinReq) {
				mockSrv.EXPECT().
					SendCoin(gomock.Any(), username, req.ToUser, req.Amount).
					Return(nil)
			},
			expStatus: http.StatusOK,
			expAns:    nil,
		},
		{
			name:         "Err No Username",
			username:     "mockuser",
			skipUsername: true,
			mockBehavior: func(username string, req sendCoinReq) {},
			expStatus:    http.StatusInternalServerError,
			expAns:       gin.H{"errors": "no username in token"},
		},
		{
			name:     "Err Service",
			username: "mockuser",
			mockBehavior: func(username string, req sendCoinReq) {
				mockSrv.EXPECT().
					SendCoin(gomock.Any(), username, req.ToUser, req.Amount).
					Return(ErrMock)
			},
			expStatus: http.StatusInternalServerError,
			expAns:    gin.H{"errors": "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username, tc.req)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if !tc.skipUsername {
				c.Set("username", tc.username)
			}

			reqMarshalled, err := json.Marshal(tc.req)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(reqMarshalled))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.SendCoins(c)

			require.Equal(t, tc.expStatus, w.Code)

			crResp, err := json.Marshal(tc.expAns)
			require.NoError(t, err)
			require.Equal(t, crResp, w.Body.Bytes())
		})
	}
}

func TestBuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := mocks.NewMockInterface(ctrl)
	handler := NewController(mockSrv)

	testCases := []struct {
		name         string
		username     string
		skipUsername bool
		itemName     string
		skipItem     bool
		mockBehavior func(username, item string)
		expStatus    int
		expAns       interface{}
	}{
		{
			name:     "OK",
			username: "mockuser",
			itemName: "mockItem",
			mockBehavior: func(username string, item string) {
				mockSrv.EXPECT().
					BuyItem(gomock.Any(), username, item).
					Return(nil)
			},
			expStatus: http.StatusOK,
		},
		{
			name:         "Err No Username",
			username:     "",
			skipUsername: true,
			itemName:     "mockItem",
			mockBehavior: func(username string, item string) {},
			expStatus:    http.StatusInternalServerError,
			expAns:       gin.H{"errors": "no username in token"},
		},
		{
			name:         "Err Invalid Items",
			username:     "mockuser",
			itemName:     "invalid",
			skipItem:     true,
			mockBehavior: func(username string, item string) {},
			expStatus:    http.StatusBadRequest,
			expAns:       gin.H{"errors": "invalid item"},
		},
		{
			name:     "Err Service",
			username: "mockuser",
			itemName: "mockItem",
			mockBehavior: func(username string, item string) {
				mockSrv.EXPECT().
					BuyItem(gomock.Any(), username, item).
					Return(ErrMock)
			},
			expStatus: http.StatusInternalServerError,
			expAns:    gin.H{"errors": "internal server error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior(tc.username, tc.itemName)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if !tc.skipUsername {
				c.Set("username", tc.username)
			}
			if !tc.skipItem {
				c.Params = append(c.Params, gin.Param{Key: "item", Value: tc.itemName})
			}

			req, err := http.NewRequest("GET", "/api/buy/"+tc.itemName, nil)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handler.BuyItem(c)

			require.Equal(t, tc.expStatus, w.Code)

			crResp, err := json.Marshal(tc.expAns)
			require.NoError(t, err)
			require.Equal(t, crResp, w.Body.Bytes())
		})
	}
}
