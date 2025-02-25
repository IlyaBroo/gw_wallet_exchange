package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gw-currency-wallet/internal/storages"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) ErrorCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) InfoCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) WarnCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) DebugCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) FatalCtx(ctx context.Context, msg string, err error) {
	m.Called(ctx, msg, err)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CheckUser(username, email string, ctx context.Context) (bool, error) {
	args := m.Called(username, email, ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) AddUser(req storages.RegisterRequest, ctx context.Context) error {
	args := m.Called(req, ctx)
	return args.Error(0)
}

func (m *MockRepository) GetUser(username string, ctx context.Context) (storages.User, error) {
	args := m.Called(username, ctx)
	return args.Get(0).(storages.User), args.Error(1)
}

func (m *MockRepository) ExchangeForCurrency(ctx context.Context, from, to string, amount decimal.Decimal, kurs float32, user_id int) (map[string]decimal.Decimal, error) {
	args := m.Called(ctx, from, to, amount, kurs, user_id)
	return args.Get(0).(map[string]decimal.Decimal), args.Error(1)
}

func (m *MockRepository) Deposit(user_id int, amount decimal.Decimal, currency string, ctx context.Context) error {
	args := m.Called(user_id, amount, currency, ctx)
	return args.Error(0)
}

func (m *MockRepository) Withdraw(user_id int, amount decimal.Decimal, currency string, ctx context.Context) error {
	args := m.Called(user_id, amount, currency, ctx)
	return args.Error(0)
}

func (m *MockRepository) GetBalance(user_id int, ctx context.Context) (storages.Balance, error) {
	args := m.Called(user_id, ctx)
	return args.Get(0).(storages.Balance), args.Error(1)
}

func (m *MockRepository) Close() {}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		input          storages.RegisterRequest
		mockCheckUser  func(m *MockRepository)
		mockAddUser    func(m *MockRepository)
		mockInfoCtx    func(m *MockLogger)
		mockErrorCtx   func(m *MockLogger)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful registration",
			input: storages.RegisterRequest{
				Username: "newuser",
				Email:    "newuser@example.com",
				Password: "password123",
			},
			mockCheckUser: func(m *MockRepository) {
				m.On("CheckUser", "newuser", "newuser@example.com", mock.Anything).Return(false, nil)
			},
			mockAddUser: func(m *MockRepository) {
				m.On("AddUser", mock.Anything, mock.Anything).Return(nil)
			},
			mockInfoCtx: func(m *MockLogger) {
				m.On("InfoCtx", mock.Anything, "User newuser registered successfully").Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"User newuser registered successfully"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockLogger := new(MockLogger)

			if tt.mockCheckUser != nil {
				tt.mockCheckUser(mockRepo)
			}
			if tt.mockAddUser != nil {
				tt.mockAddUser(mockRepo)
			}
			if tt.mockInfoCtx != nil {
				tt.mockInfoCtx(mockLogger)
			}
			if tt.mockErrorCtx != nil {
				tt.mockErrorCtx(mockLogger)
			}

			s := &ServerWallet{
				db: mockRepo,
				lg: mockLogger,
			}

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			s.RegisterUser(w, req)

			res := w.Result()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
