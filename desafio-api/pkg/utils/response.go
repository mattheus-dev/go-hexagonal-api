package utils
import "github.com/gin-gonic/gin"
type ErrorResponse struct {
    Error string `json:"error"`
}
func RespondWithError(c *gin.Context, code int, message string) {
    c.JSON(code, ErrorResponse{Error: message})
}
