package http
import (
	"net/http"
	"strings"
	"desafio-api/internal/application/service"
	"github.com/gin-gonic/gin"
)
func AuthMiddleware(userService service.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithError(c, http.StatusUnauthorized, "Token de autenticação ausente ou inválido")
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := userService.ValidateToken(tokenString)
		if err != nil {
			respondWithError(c, http.StatusUnauthorized, "Token de autenticação inválido ou expirado")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
