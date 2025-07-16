package repository
import (
    "context"
    "desafio-api/internal/domain"
)
type ItemRepository interface {
    Save(ctx context.Context, item *domain.Item) error
    Update(ctx context.Context, item *domain.Item) error
    FindByID(ctx context.Context, id int64) (*domain.Item, error)
    FindAll(ctx context.Context, status string, limit, offset int) ([]*domain.Item, int, error)
    Delete(ctx context.Context, id int64) error
    ExistsByCode(ctx context.Context, code string, excludeID int64) (bool, error)
}
