package http
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
type MockUserService struct {
	mock.Mock
}
func (m *MockUserService) Register(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}
func (m *MockUserService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.JWTClaims), args.Error(1)
}
func (m *MockUserService) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserService) GetRepository() service.UserRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(service.UserRepository)
}
func setupTest() (*gin.Engine, *MockUserService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockUserService)
	handler := NewAuthHandler(mockService)
	router := gin.Default()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	return router, mockService
}
func TestRegister_Success(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Register", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Username == "newuser" && user.Password == "password123"
	})).Return(nil)
	reqBody := map[string]string{
		"username": "newuser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}
func TestRegister_InvalidRequest(t *testing.T) {
	router, _ := setupTest()
	reqBody := map[string]string{
		"username": "newuser",
		"password": "short",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestRegister_UserAlreadyExists(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Register", mock.Anything, mock.Anything).Return(domain.ErrDuplicateUsername)
	reqBody := map[string]string{
		"username": "existinguser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}
func TestRegister_ServerError(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Register", mock.Anything, mock.Anything).Return(errors.New("db error"))
	reqBody := map[string]string{
		"username": "newuser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
func TestLogin_Success(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Login", mock.Anything, "testuser", "password123").Return("jwt-token-123", nil)
	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "jwt-token-123", response["token"])
	mockService.AssertExpectations(t)
}
func TestLogin_InvalidCredentials(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Login", mock.Anything, "wronguser", "wrongpass").Return("", domain.ErrInvalidCredentials)
	reqBody := map[string]string{
		"username": "wronguser",
		"password": "wrongpass",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}
func TestLogin_InvalidRequest(t *testing.T) {
	router, _ := setupTest()
	reqBody := map[string]string{
		"username": "", 
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestLogin_ServerError(t *testing.T) {
	router, mockService := setupTest()
	mockService.On("Login", mock.Anything, "testuser", "password123").Return("", errors.New("database error"))
	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
