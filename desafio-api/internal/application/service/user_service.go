package service
import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"desafio-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id int) (*domain.User, error)
}
type UserService struct {
	userRepo  UserRepository
	jwtSecret string
}
func NewUserService(userRepo UserRepository) *UserService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-jwt-key"
		log.Printf("[WARN] JWT Secret not found in environment variables, using default secret (NOT SAFE FOR PRODUCTION)")
	}
	log.Printf("[INFO] JWT Secret configured with length: %d", len(secret))
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: secret,
	}
}
func (s *UserService) Register(ctx context.Context, user *domain.User) error {
	log.Printf("[DEBUG] UserService.Register: Registering user: %s", user.Username)
	if err := user.Validate(); err != nil {
		log.Printf("[ERROR] UserService.Register: User validation failed: %v", err)
		return err
	}
	if err := user.HashPassword(); err != nil {
		log.Printf("[ERROR] UserService.Register: Failed to hash password: %v", err)
		return err
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("[ERROR] UserService.Register: Failed to create user in repository: %v", err)
		return err
	}
	log.Printf("[INFO] UserService.Register: User registered successfully: %s (ID: %d)", user.Username, user.ID)
	return nil
}
func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	log.Printf("[DEBUG] UserService.Login: Login attempt for user: %s", username)
	if username == "" {
		log.Printf("[ERROR] UserService.Login: Nome de usuário vazio")
		return "", domain.ErrInvalidCredentials
	}
	if password == "" {
		log.Printf("[ERROR] UserService.Login: Senha vazia")
		return "", domain.ErrInvalidCredentials
	}
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		log.Printf("[ERROR] UserService.Login: User not found: %s, error: %v", username, err)
		return "", domain.ErrInvalidCredentials
	}
	if !user.ComparePassword(password) {
		log.Printf("[ERROR] UserService.Login: Invalid password for user: %s", username)
		return "", domain.ErrInvalidCredentials
	}
	if user.ID <= 0 {
		log.Printf("[ERROR] UserService.Login: User has invalid ID: %d", user.ID)
		return "", fmt.Errorf("usuário com ID inválido")
	}
	token, err := s.generateToken(user)
	if err != nil {
		log.Printf("[ERROR] UserService.Login: Failed to generate token: %v", err)
		return "", err
	}
	log.Printf("[INFO] UserService.Login: Login successful for user: %s", username)
	return token, nil
}
func (s *UserService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	if len(tokenString) < 10 {
		log.Printf("[ERROR] UserService.ValidateToken: Token too short: %s", tokenString)
		return nil, domain.ErrInvalidToken
	}
	log.Printf("[DEBUG] UserService.ValidateToken: Validating token: %s...", tokenString[:10])
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[ERROR] UserService.ValidateToken: Unexpected signing method: %v", token.Header["alg"])
			return nil, domain.ErrInvalidToken
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		log.Printf("[ERROR] UserService.ValidateToken: Token validation failed: %v", err)
		return nil, domain.ErrInvalidToken
	}
	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		log.Printf("[ERROR] UserService.ValidateToken: Invalid token claims")
		return nil, domain.ErrInvalidToken
	}
	log.Printf("[DEBUG] UserService.ValidateToken: Token validated successfully for user ID: %d", claims.UserID)
	return claims, nil
}
func (s *UserService) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	log.Printf("[DEBUG] UserService.GetUserByID: Getting user with ID: %d", id)
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		log.Printf("[ERROR] UserService.GetUserByID: Failed to find user with ID %d: %v", id, err)
		return nil, err
	}
	return user, nil
}
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	log.Printf("[DEBUG] UserService.GetUserByUsername: Buscando usuário pelo username: %s", username)
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		log.Printf("[ERROR] UserService.GetUserByUsername: Erro ao buscar usuário: %v", err)
		return nil, err
	}
	return user, nil
}
func (s *UserService) GetRepository() interface{} {
	return s.userRepo
}
func (s *UserService) GetJWTSecret() string {
	return s.jwtSecret
}
func (s *UserService) generateToken(user *domain.User) (string, error) {
	log.Printf("[DEBUG] UserService.generateToken: Generating token for user ID: %d", user.ID)
	now := time.Now()
	expirationTime := now.Add(1 * time.Hour)
	if user.ID <= 0 {
		log.Printf("[ERROR] UserService.generateToken: Tentando gerar token para usuário com ID inválido: %d", user.ID)
		return "", fmt.Errorf("ID de usuário inválido")
	}
	claims := &domain.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if s.jwtSecret == "" {
		log.Printf("[ERROR] UserService.generateToken: JWT secret está vazio")
		s.jwtSecret = "your-super-secret-jwt-key" 
	}
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Printf("[ERROR] UserService.generateToken: Failed to sign token: %v", err)
		return "", err
	}
	log.Printf("[INFO] UserService.generateToken: Token generated successfully for user: %s", user.Username)
	return tokenString, nil
}
