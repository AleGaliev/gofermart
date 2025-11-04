package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AleGaliev/gofermart/internal/mocks"
	model "github.com/AleGaliev/gofermart/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMyHandler_ServeHTTPRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)
	handler := &MyHandler{
		storage:    mockStorage,
		jwtManager: mockJWT,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		contentType    string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "Successful registration",
			requestBody: model.User{
				Login:    "testuser",
				Password: "testpass",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockStorage.EXPECT().Register(gomock.Any()).Return(nil)
				mockJWT.EXPECT().IssueJWT(gomock.Any()).Return("test-token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid content type",
			requestBody: model.User{
				Login:    "testuser",
				Password: "testpass",
			},
			contentType:    "text/plain",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User already exists",
			requestBody: model.User{
				Login:    "existinguser",
				Password: "testpass",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockStorage.EXPECT().Register(gomock.Any()).Return(errors.New("user already exists"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "JWT issue error",
			requestBody: model.User{
				Login:    "testuser",
				Password: "testpass",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockStorage.EXPECT().Register(gomock.Any()).Return(nil)
				mockJWT.EXPECT().IssueJWT(gomock.Any()).Return("", errors.New("jwt error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTPRegister(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestMyHandler_ServeHTTPLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)
	handler := &MyHandler{
		storage:    mockStorage,
		jwtManager: mockJWT,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		contentType    string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "Successful login",
			requestBody: model.User{
				Login:    "testuser",
				Password: "testpass",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockStorage.EXPECT().Login(gomock.Any()).Return(nil)
				mockJWT.EXPECT().IssueJWT(gomock.Any()).Return("test-token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid content type",
			requestBody: model.User{
				Login:    "testuser",
				Password: "testpass",
			},
			contentType:    "text/plain",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid credentials",
			requestBody: model.User{
				Login:    "wronguser",
				Password: "wrongpass",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockStorage.EXPECT().Login(gomock.Any()).Return(errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTPLogin(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestMyHandler_ServeHTTPBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)
	handler := &MyHandler{
		storage:    mockStorage,
		jwtManager: mockJWT,
	}

	tests := []struct {
		name           string
		userLogin      string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name:      "Successful balance retrieval",
			userLogin: "testuser",
			setupMocks: func() {
				balance := model.UserBalanceOut{
					Current:   100.5,
					Withdrawn: 50.0,
				}
				mockStorage.EXPECT().GetUserBalance("testuser").Return(balance, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Storage error",
			userLogin: "testuser",
			setupMocks: func() {
				mockStorage.EXPECT().GetUserBalance("testuser").Return(model.UserBalanceOut{}, errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest("GET", "/api/user/balance", nil)
			req.Header.Set("X-User-Login", tt.userLogin)

			rr := httptest.NewRecorder()
			handler.ServeHTTPBalance(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response model.UserBalance
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				expectedCurrent, _ := decimal.NewFromString("100.5")
				expectedWithdrawn, _ := decimal.NewFromString("50.0")

				assert.True(t, expectedCurrent.Equal(response.Current))
				assert.True(t, expectedWithdrawn.Equal(response.Withdrawn))
			}
		})
	}
}

func TestMyHandler_ServeHTTPOrdersNumders(t *testing.T) {
	tests := []struct {
		name           string
		userLogin      string
		orderNumber    string
		setupMocks     func(*mocks.MockStorage, *mocks.MockJWTManager)
		expectedStatus int
	}{
		{
			name:        "Successful order upload - new order",
			userLogin:   "testuser",
			orderNumber: "12345678903",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetOrder("12345678903").Return("", nil)
				ms.EXPECT().UploadOrder("testuser", "12345678903").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "Order already uploaded by this user",
			userLogin:   "testuser",
			orderNumber: "12345678903",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetOrder("12345678903").Return("testuser", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Order taken by another user",
			userLogin:   "testuser",
			orderNumber: "12345678903",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetOrder("12345678903").Return("otheruser", nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Empty order number",
			userLogin:      "testuser",
			orderNumber:    "",
			setupMocks:     func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid order number",
			userLogin:      "testuser",
			orderNumber:    "123",
			setupMocks:     func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "Error checking order existence",
			userLogin:   "testuser",
			orderNumber: "12345678903",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetOrder("12345678903").Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Error uploading order",
			userLogin:   "testuser",
			orderNumber: "12345678903",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetOrder("12345678903").Return("", nil)
				ms.EXPECT().UploadOrder("testuser", "12345678903").Return(errors.New("upload error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockStorage(ctrl)
			mockJWT := mocks.NewMockJWTManager(ctrl)
			handler := &MyHandler{
				storage:    mockStorage,
				jwtManager: mockJWT,
			}

			tt.setupMocks(mockStorage, mockJWT)

			req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set("X-User-Login", tt.userLogin)

			rr := httptest.NewRecorder()
			handler.ServeHTTPOrdersNumders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestMyHandler_ServeHTTPOrdersInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)
	handler := &MyHandler{
		storage:    mockStorage,
		jwtManager: mockJWT,
	}

	tests := []struct {
		name           string
		userLogin      string
		setupMocks     func()
		expectedStatus int
		checkResponse  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:      "Successful orders retrieval",
			userLogin: "testuser",
			setupMocks: func() {
				orders := []model.OrderOut{
					{
						OrderNumber: "12345678903",
						Status:      "PROCESSED",
						Accrual:     100.0,
						UploadedAt:  time.Now(),
					},
				}
				mockStorage.EXPECT().GetOrdersUser("testuser").Return(orders, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []model.Order
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, "12345678903", response[0].OrderNumber)
				assert.Equal(t, "PROCESSED", response[0].Status)
				expectedAccrual := decimal.NewFromFloat(100.0)
				assert.True(t, expectedAccrual.Equal(response[0].Accrual))
			},
		},
		{
			name:      "No orders found",
			userLogin: "testuser",
			setupMocks: func() {
				mockStorage.EXPECT().GetOrdersUser("testuser").Return([]model.OrderOut{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  nil,
		},
		{
			name:      "Storage error",
			userLogin: "testuser",
			setupMocks: func() {
				mockStorage.EXPECT().GetOrdersUser("testuser").Return([]model.OrderOut{}, errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest("GET", "/api/user/orders", nil)
			req.Header.Set("X-User-Login", tt.userLogin)

			rr := httptest.NewRecorder()
			handler.ServeHTTPOrdersInfo(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestMyHandler_ServeHTTPBalanceWithDraw(t *testing.T) {
	tests := []struct {
		name           string
		userLogin      string
		withdraw       model.WithdrawOut
		contentType    string
		requestBody    string
		setupMocks     func(*mocks.MockStorage, *mocks.MockJWTManager)
		expectedStatus int
	}{
		{
			name:      "Successful withdraw",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         50.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				userBalance := model.UserBalanceOut{
					Current:   100.0,
					Withdrawn: 0.0,
				}
				ms.EXPECT().GetUserBalance("testuser").Return(userBalance, nil)

				ms.EXPECT().GetOrder("12345678903").Return("", nil)

				ms.EXPECT().UploadOrderWithdraw("testuser", gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid content type",
			userLogin:      "testuser",
			withdraw:       model.WithdrawOut{},
			contentType:    "text/plain",
			setupMocks:     func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON body",
			userLogin:      "testuser",
			contentType:    "application/json",
			requestBody:    "invalid json",
			setupMocks:     func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid order number",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "123",
				Sum:         50.0,
			},
			contentType:    "application/json",
			setupMocks:     func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:      "Insufficient funds",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         1000.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				userBalance := model.UserBalanceOut{
					Current:   100.0,
					Withdrawn: 0.0,
				}
				ms.EXPECT().GetUserBalance("testuser").Return(userBalance, nil)
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:      "Order already exists for withdraw",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         50.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				userBalance := model.UserBalanceOut{
					Current:   100.0,
					Withdrawn: 0.0,
				}
				ms.EXPECT().GetUserBalance("testuser").Return(userBalance, nil)
				ms.EXPECT().GetOrder("12345678903").Return("someuser", nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:      "Error getting user balance",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         50.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				ms.EXPECT().GetUserBalance("testuser").Return(model.UserBalanceOut{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Error checking order existence",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         50.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				userBalance := model.UserBalanceOut{
					Current:   100.0,
					Withdrawn: 0.0,
				}
				ms.EXPECT().GetUserBalance("testuser").Return(userBalance, nil)
				ms.EXPECT().GetOrder("12345678903").Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Error uploading withdraw",
			userLogin: "testuser",
			withdraw: model.WithdrawOut{
				OrderNumber: "12345678903",
				Sum:         50.0,
			},
			contentType: "application/json",
			setupMocks: func(ms *mocks.MockStorage, mj *mocks.MockJWTManager) {
				userBalance := model.UserBalanceOut{
					Current:   100.0,
					Withdrawn: 0.0,
				}
				ms.EXPECT().GetUserBalance("testuser").Return(userBalance, nil)
				ms.EXPECT().GetOrder("12345678903").Return("", nil)
				ms.EXPECT().UploadOrderWithdraw("testuser", gomock.Any()).Return(errors.New("upload error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockStorage(ctrl)
			mockJWT := mocks.NewMockJWTManager(ctrl)
			handler := &MyHandler{
				storage:    mockStorage,
				jwtManager: mockJWT,
			}

			tt.setupMocks(mockStorage, mockJWT)

			var body []byte
			if tt.requestBody != "" {
				body = []byte(tt.requestBody)
			} else {
				body, _ = json.Marshal(tt.withdraw)
			}

			req := httptest.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			req.Header.Set("X-User-Login", tt.userLogin)

			rr := httptest.NewRecorder()
			handler.ServeHTTPBalanceWithDraw(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestMyHandler_ServeHTTPWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)
	handler := &MyHandler{
		storage:    mockStorage,
		jwtManager: mockJWT,
	}

	now := time.Now()

	tests := []struct {
		name           string
		userLogin      string
		setupMocks     func()
		expectedStatus int
		checkResponse  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:      "Successful withdrawals retrieval",
			userLogin: "testuser",
			setupMocks: func() {
				withdrawals := []model.WithdrawOut{
					{
						OrderNumber: "12345678903",
						Sum:         50.0,
						ProcessedAt: now,
					},
				}
				mockStorage.EXPECT().GetUserWithdrawals("testuser").Return(withdrawals, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []model.Withdraw
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, "12345678903", response[0].OrderNumber)
				expectedAccrual := decimal.NewFromFloat(50.0)
				assert.True(t, expectedAccrual.Equal(response[0].Sum))
			},
		},
		{
			name:      "No withdrawals found",
			userLogin: "testuser",
			setupMocks: func() {
				mockStorage.EXPECT().GetUserWithdrawals("testuser").Return([]model.WithdrawOut{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  nil,
		},
		{
			name:      "Storage error",
			userLogin: "testuser",
			setupMocks: func() {
				mockStorage.EXPECT().GetUserWithdrawals("testuser").Return([]model.WithdrawOut{}, errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
			req.Header.Set("X-User-Login", tt.userLogin)

			rr := httptest.NewRecorder()
			handler.ServeHTTPWithdrawals(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	rr := httptest.NewRecorder()

	successResponse(rr, "test-key", http.StatusOK)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Запрос обработан", response["message"])
}

func TestFailedResponse(t *testing.T) {
	rr := httptest.NewRecorder()

	failedResponse(rr, "test error", "test-key", http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "failed", response["status"])
	assert.Equal(t, "test error", response["message"])
}
