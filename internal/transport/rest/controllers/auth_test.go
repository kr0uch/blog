package controllers

import (
	"blog/internal/models/dto"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
	secret string
}

func (m *MockAuthService) RegistrateUser(user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RegistrateUserResponse), args.Error(1)
}

func (m *MockAuthService) LoginUser(user *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginUserResponse), args.Error(1)
}

func (m *MockAuthService) RefreshUserToken(token *dto.RefreshUserTokenRequest) (*dto.RefreshUserTokenResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RefreshUserTokenResponse), args.Error(1)
}

func TestAuthController_RegistrateUser(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockFunc           func(m *MockAuthService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string, responseHeader string)
	}{
		{
			name: "successful",
			requestBody: &dto.RegistrateUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
				Role:     consts.AuthorRole,
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(&dto.RegistrateUserResponse{
						Message:      "message",
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string, responseHeader string) {
				var response dto.RegistrateUserResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
				assert.NotEmpty(t, responseHeader)
			},
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			requestBody: &dto.RegistrateUserRequest{
				Email:    "test_exists@yandex.ru",
				Password: "password",
				Role:     consts.AuthorRole,
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrUserAlreadyExists)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid email",
			requestBody: &dto.RegistrateUserRequest{
				Email:    "",
				Password: "password",
				Role:     consts.AuthorRole,
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrInvalidEmail)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid role",
			requestBody: &dto.RegistrateUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
				Role:     "invalidRole",
			},
			expectedStatusCode: http.StatusForbidden,
			mockFunc: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrInvalidRole)
			},
		},
		{
			name: "internal server error",
			requestBody: &dto.RegistrateUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
				Role:     consts.AuthorRole,
			},
			expectedStatusCode: http.StatusForbidden,
			mockFunc: func(m *MockAuthService) {
				m.On("RegistrateUser", mock.AnythingOfType("*dto.RegistrateUserRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAuthService := &MockAuthService{secret: "test"}
			if test.mockFunc != nil {
				test.mockFunc(mockAuthService)
			}

			controller := NewAuthController(mockAuthService)

			req := &http.Request{}

			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			controller.RegistrateUser(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String(), rr.Header().Get("Authorization"))
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_LoginUser(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockFunc           func(m *MockAuthService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string, responseHeader string)
	}{
		{
			name: "successful",
			requestBody: &dto.LoginUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(&dto.LoginUserResponse{
						Message:      "message",
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string, responseHeader string) {
				var response dto.LoginUserResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
				assert.NotEmpty(t, response.RefreshToken)
				assert.NotEmpty(t, responseHeader)
			},
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			requestBody: &dto.LoginUserRequest{
				Email:    "test@",
				Password: "password",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrInvalidEmail)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid email or password",
			requestBody: &dto.LoginUserRequest{
				Email:    "",
				Password: "",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrInvalidEmailOrPassword)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid user id",
			requestBody: &dto.LoginUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrInvalidUserId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			requestBody: &dto.LoginUserRequest{
				Email:    "test@yandex.ru",
				Password: "password",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("LoginUser", mock.AnythingOfType("*dto.LoginUserRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAuthService := &MockAuthService{secret: "test"}
			if test.mockFunc != nil {
				test.mockFunc(mockAuthService)
			}
			controller := NewAuthController(mockAuthService)

			req := &http.Request{}
			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
			}

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			controller.LoginUser(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String(), rr.Header().Get("Authorization"))
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthController_RefreshUserToken(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockFunc           func(m *MockAuthService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string, responseHeader string)
	}{
		{
			name: "successful",
			requestBody: &dto.RefreshUserTokenRequest{
				RefreshToken: "refresh_token",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RefreshUserToken", mock.AnythingOfType("*dto.RefreshUserTokenRequest")).
					Return(&dto.RefreshUserTokenResponse{
						Message:      "message",
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string, responseHeader string) {
				var response dto.RefreshUserTokenResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
				assert.NotEmpty(t, responseHeader)
			},
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid refresh token",
			requestBody: &dto.RefreshUserTokenRequest{
				RefreshToken: "",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RefreshUserToken", mock.AnythingOfType("*dto.RefreshUserTokenRequest")).
					Return(nil, errors.ErrInvalidRefreshToken)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "internal server error",
			requestBody: &dto.RefreshUserTokenRequest{
				RefreshToken: "refresh_token",
			},
			mockFunc: func(m *MockAuthService) {
				m.On("RefreshUserToken", mock.AnythingOfType("*dto.RefreshUserTokenRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockAuthService := &MockAuthService{secret: "test"}
			if test.mockFunc != nil {
				test.mockFunc(mockAuthService)
			}
			controller := NewAuthController(mockAuthService)

			req := &http.Request{}
			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", nil)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			controller.RefreshUserToken(rr, req)

			assert.Equal(t, test.expectedStatusCode, rr.Code)
			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String(), rr.Header().Get("Authorization"))
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}
