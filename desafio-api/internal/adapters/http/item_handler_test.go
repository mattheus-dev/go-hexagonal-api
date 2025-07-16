package http
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"desafio-api/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
type ItemServiceInterface interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id int64) (*domain.Item, error)
	Update(ctx context.Context, id int64, item *domain.Item) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, status string, page, perPage int) ([]*domain.Item, int, error)
}
type MockItemService struct {
	mock.Mock
}
func (m *MockItemService) Create(ctx context.Context, item *domain.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}
func (m *MockItemService) GetByID(ctx context.Context, id int64) (*domain.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Item), args.Error(1)
}
func (m *MockItemService) Update(ctx context.Context, id int64, item *domain.Item) error {
	args := m.Called(ctx, id, item)
	return args.Error(0)
}
func (m *MockItemService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockItemService) List(ctx context.Context, status string, page, perPage int) ([]*domain.Item, int, error) {
	args := m.Called(ctx, status, page, perPage)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Item), args.Int(1), args.Error(2)
}
func NewItemHandlerWithInterface(service ItemServiceInterface) *ItemHandler {
	return &ItemHandler{itemService: service}
}
func setupItemTest() (*gin.Engine, *MockItemService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockItemService)
	handler := NewItemHandlerWithInterface(mockService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", 1)
		c.Set("username", "testuser")
		c.Next()
	})
	router.POST("/items", handler.Create)
	router.GET("/items", handler.List)
	router.GET("/items/:id", handler.GetByID)
	router.PUT("/items/:id", handler.Update)
	router.DELETE("/items/:id", handler.Delete)
	return router, mockService
}
func createTestItem() *domain.Item {
	now := time.Now()
	return &domain.Item{
		ID:          1,
		Code:        "TEST001",
		Title:       "Item de Teste",
		Description: "Descrição do item de teste",
		Price:       1500,
		Stock:       10,
		Status:      "ACTIVE",
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   1, 
		UpdatedBy:   1, 
	}
}
func TestCreate_Success(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("Create", mock.Anything, mock.MatchedBy(func(item *domain.Item) bool {
		return item.Code == "TEST001" && item.CreatedBy == 1 && item.UpdatedBy == 1
	})).Return(nil).Run(func(args mock.Arguments) {
		item := args.Get(1).(*domain.Item)
		item.ID = 1
	})
	reqBody := map[string]interface{}{
		"code":        "TEST001",
		"title":       "Item de Teste",
		"description": "Descrição do item de teste",
		"price":       1500,
		"stock":       10,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["created_by"])
	assert.Equal(t, float64(1), response["updated_by"])
	mockService.AssertExpectations(t)
}
func TestCreate_InvalidRequest(t *testing.T) {
	router, _ := setupItemTest()
	reqBody := map[string]interface{}{
		"title":       "Item de Teste",
		"description": "Descrição do item de teste",
		"price":       1500,
		"stock":       10,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestCreate_DuplicateCode(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("Create", mock.Anything, mock.Anything).Return(domain.ErrDuplicateCode)
	reqBody := map[string]interface{}{
		"code":        "DUPLICATE",
		"title":       "Item com Código Duplicado",
		"description": "Este código já existe",
		"price":       1500,
		"stock":       10,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}
func TestCreate_ServerError(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
	reqBody := map[string]interface{}{
		"code":        "TEST001",
		"title":       "Item de Teste",
		"description": "Descrição do item de teste",
		"price":       1500,
		"stock":       10,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
func TestUpdate_Success(t *testing.T) {
	router, mockService := setupItemTest()
	existingItem := createTestItem()
	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingItem, nil)
	mockService.On("Update", mock.Anything, int64(1), mock.MatchedBy(func(item *domain.Item) bool {
		return item.ID == 1 && 
			   item.Title == "Item Atualizado" && 
			   item.CreatedBy == 1 && 
			   item.UpdatedBy == 1    
	})).Return(nil)
	reqBody := map[string]interface{}{
		"code":        "TEST001",
		"title":       "Item Atualizado",
		"description": "Descrição atualizada",
		"price":       2000,
		"stock":       15,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/items/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["created_by"]) 
	assert.Equal(t, float64(1), response["updated_by"]) 
	mockService.AssertExpectations(t)
}
func TestUpdate_ItemNotFound(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.ErrItemNotFound)
	reqBody := map[string]interface{}{
		"code":        "TEST999",
		"title":       "Item Inexistente",
		"description": "Este item não existe",
		"price":       2000,
		"stock":       15,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/items/999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
func TestUpdate_InvalidRequest(t *testing.T) {
	router, _ := setupItemTest()
	reqBody := map[string]interface{}{
		"code":        "", 
		"title":       "Item Atualizado",
		"description": "Descrição atualizada",
		"price":       2000,
		"stock":       15,
	}
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/items/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestGetByID_Success(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("GetByID", mock.Anything, int64(1)).Return(createTestItem(), nil)
	req, _ := http.NewRequest("GET", "/items/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["created_by"])
	assert.Equal(t, float64(1), response["updated_by"])
	mockService.AssertExpectations(t)
}
func TestGetByID_NotFound(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.ErrItemNotFound)
	req, _ := http.NewRequest("GET", "/items/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
func TestList_Success(t *testing.T) {
	router, mockService := setupItemTest()
	items := []*domain.Item{createTestItem()}
	totalPages := 1
	mockService.On("List", mock.Anything, "", 1, 10).Return(items, totalPages, nil)
	req, _ := http.NewRequest("GET", "/items?page=1&per_page=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["totalPages"])
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
	item := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), item["created_by"])
	assert.Equal(t, float64(1), item["updated_by"])
	mockService.AssertExpectations(t)
}
func TestDelete_Success(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)
	req, _ := http.NewRequest("DELETE", "/items/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestDelete_NotFound(t *testing.T) {
	router, mockService := setupItemTest()
	mockService.On("Delete", mock.Anything, int64(999)).Return(domain.ErrItemNotFound)
	req, _ := http.NewRequest("DELETE", "/items/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
