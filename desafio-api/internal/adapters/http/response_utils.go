package http
import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
)
func RespondWithError(c *gin.Context, status int, message string) {
	log.Printf("[DEBUG] Respondendo com erro HTTP %d: %s", status, message)
	c.JSON(status, gin.H{
		"error": message,
		"status": status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
