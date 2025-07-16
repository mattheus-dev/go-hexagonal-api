package http
import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)
func TestErrorMiddleware_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware())
	router.GET("/no-panic", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	req, _ := http.NewRequest("GET", "/no-panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}
func TestErrorMiddleware_WithPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware())
	router.GET("/panic-error", func(c *gin.Context) {
		panic(errors.New("test error"))
	})
	req, _ := http.NewRequest("GET", "/panic-error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "erro interno")
}
func TestErrorMiddleware_WithPanicString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware())
	router.GET("/panic-string", func(c *gin.Context) {
		panic("test panic string")
	})
	req, _ := http.NewRequest("GET", "/panic-string", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "erro interno")
}
func TestErrorMiddleware_WithPanicOther(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware())
	router.GET("/panic-other", func(c *gin.Context) {
		panic(123) 
	})
	req, _ := http.NewRequest("GET", "/panic-other", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "erro interno")
}
