package http
import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		errorID := uuid.New().String()
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				log.Printf("[CRITICAL] Panic recovered [ID: %s]:\nError: %v\nStack Trace:\n%s", 
					errorID, r, stack)
				var errorMessage string
				switch v := r.(type) {
				case error:
					errorMessage = v.Error()
				case string:
					errorMessage = v
				default:
					errorMessage = fmt.Sprintf("%v", r)
				}
				clientIP := c.ClientIP()
				method := c.Request.Method
				path := c.Request.URL.Path
				userID, exists := c.Get("userID")
				errorLog := map[string]interface{}{
					"error_id":  errorID,
					"message":   errorMessage,
					"method":    method,
					"path":      path,
					"remote_ip": clientIP,
					"status":    http.StatusInternalServerError,
				}
				if exists {
					errorLog["user_id"] = userID
				}
				log.Printf("[CRITICAL] Erro do cliente [ID: %s]: %v", errorID, errorLog)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Ocorreu um erro interno no servidor",
					"error_id":   errorID,
					"status":     http.StatusInternalServerError,
					"timestamp":  time.Now().Format(time.RFC3339),
					"message":    errorMessage, 
				})
			}
		}()
		c.Next()
		if len(c.Errors) > 0 {
			errorID := uuid.New().String()
			for _, err := range c.Errors {
				log.Printf("[ERROR] Gin error [ID: %s]: %v", errorID, err)
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":     "Ocorreu um erro não tratado",
						"error_id":  errorID,
						"timestamp": time.Now().Format(time.RFC3339),
					})
					break
				}
			}
		}
	}
}
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestPath := c.Request.URL.Path
		requestMethod := c.Request.Method
		clientIP := c.ClientIP()
		log.Printf("[INFO] Request: %s %s from %s", requestMethod, requestPath, clientIP)
		c.Next()
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		logLevel := "[INFO]"
		if statusCode >= 400 && statusCode < 500 {
			logLevel = "[WARN]"
		} else if statusCode >= 500 {
			logLevel = "[ERROR]"
		}
		log.Printf("%s Response: %s %s | Status: %d | Latency: %v", 
			logLevel, requestMethod, requestPath, statusCode, latency)
	}
}
