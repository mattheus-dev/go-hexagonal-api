package service
import (
	"context"
	"desafio-api/internal/domain"
)
type UserServiceInterface interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetJWTSecret() string
	GetRepository() interface{}
}
var _ UserServiceInterface = (*UserService)(nil)
