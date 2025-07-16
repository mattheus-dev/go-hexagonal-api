package http
import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
type MockUserServiceForAuth struct {
	mock.Mock
}
func (m *MockUserServiceForAuth) Register(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserServiceForAuth) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}
func (m *MockUserServiceForAuth) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.JWTClaims), args.Error(1)
}
func (m *MockUserServiceForAuth) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserServiceForAuth) GetRepository() service.UserRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(service.UserRepository)
}
func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUserService := new(MockUserServiceForAuth)
	middleware := AuthMiddleware(mockUserService)
	router := gin.New()
	router.GET("/protected", middleware, func(c *gin.Context) {
		userID, exists := c.Get("userID")
		username, usernameExists := c.Get("username")
		if exists && usernameExists {
			c.JSON(http.StatusOK, gin.H{
				"userID":   userID,
				"username": username,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Context values not set"})
		}
	})
	t.Run("Valid Token", func(t *testing.T) {
		claims := &domain.JWTClaims{
			UserID:   1,
			Username: "testuser",
		}
		mockUserService.On("ValidateToken", "valid-token").Return(claims, nil).Once()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["userID"])
		assert.Equal(t, "testuser", response["username"])
		mockUserService.AssertExpectations(t)
	})
	t.Run("Missing Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("Invalid Token", func(t *testing.T) {
		mockUserService.On("ValidateToken", "invalid-token").Return(nil, domain.ErrInvalidToken).Once()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})
}
