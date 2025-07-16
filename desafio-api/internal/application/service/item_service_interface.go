package service
import (
	"context"
	"desafio-api/internal/domain"
)
type ItemServiceInterface interface {
	Create(ctx context.Context, item *domain.Item) error
	GetByID(ctx context.Context, id int64) (*domain.Item, error)
	Update(ctx context.Context, id int64, item *domain.Item) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, status string, page, perPage int) ([]*domain.Item, int, error)
}
var _ ItemServiceInterface = (*ItemService)(nil)
