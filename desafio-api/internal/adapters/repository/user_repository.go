package repository
import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"desafio-api/internal/domain"
	"github.com/jmoiron/sqlx"
)
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id int) (*domain.User, error)
}
var _ UserRepositoryInterface = (*UserRepository)(nil)
type UserRepository struct {
	db *sqlx.DB
}
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, password)
		VALUES (?, ?)
	`
	existingUser, err := r.FindByUsername(ctx, user.Username)
	if err == nil && existingUser != nil {
		log.Printf("[DEBUG] UserRepository.Create: Username já existe: %s", user.Username)
		return domain.ErrDuplicateUsername
	} else if err != nil && err != domain.ErrUserNotFound {
		log.Printf("[ERROR] UserRepository.Create: Erro ao verificar existência do usuário: %v", err)
		return err
	}
	log.Printf("[DEBUG] UserRepository.Create: Inserindo novo usuário: %s", user.Username)
	result, err := r.db.ExecContext(ctx, query, user.Username, user.Password)
	if err != nil {
		if isDuplicateKeyError(err) {
			log.Printf("[ERROR] UserRepository.Create: Erro de chave duplicada: %v", err)
			return domain.ErrDuplicateUsername
		}
		log.Printf("[ERROR] UserRepository.Create: Erro ao criar usuário: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("[ERROR] UserRepository.Create: Erro ao obter ID do usuário inserido: %v", err)
		return err
	}
	user.ID = int(id)
	log.Printf("[INFO] UserRepository.Create: Usuário criado com sucesso: ID=%d, Username=%s", user.ID, user.Username)
	return nil
}
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, username, password, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "for key")
}
