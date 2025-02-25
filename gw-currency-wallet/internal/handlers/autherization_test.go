package handlers

import (
	"bytes"
	"encoding/json"
	"gw-currency-wallet/internal/storages"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAutherisation(t *testing.T) {
	tests := []struct {
		name           string
		input          LoginRequest
		mockGetUser    func(m *MockRepository)
		mockErrorCtx   func(m *MockLogger)
		mockInfoCtx    func(m *MockLogger)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful login",
			input: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockGetUser: func(m *MockRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				m.On("GetUser", "testuser", mock.Anything).Return(storages.User{Id: 1, Username: "testuser", Password: string(hashedPassword)}, nil)
			},
			mockInfoCtx: func(m *MockLogger) {
				m.On("InfoCtx", mock.Anything, "User testuser logged in successfully").Return(nil)
			},
			mockErrorCtx:   func(m *MockLogger) {},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка мока
			mockDB := new(MockRepository)
			mockLogger := new(MockLogger)

			tt.mockGetUser(mockDB)
			tt.mockInfoCtx(mockLogger)
			tt.mockErrorCtx(mockLogger)

			s := &ServerWallet{
				db: mockDB,
				lg: mockLogger,
			}

			// Подготовка запроса
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			// Вызов обработчика
			s.Autherisation(w, req)

			// Проверка статуса
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Проверка тела ответа
			bodyResp, _ := ioutil.ReadAll(resp.Body)
			if tt.expectedBody == `{"token":` {
				assert.Contains(t, string(bodyResp), "token") // Проверяем, что токен присутствует
			} else {
				assert.JSONEq(t, tt.expectedBody, string(bodyResp))
			}

			// Проверка вызовов моков
			mockDB.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
