package domain
import (
	"github.com/golang-jwt/jwt/v5"
)
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
