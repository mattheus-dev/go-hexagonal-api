package service
import (
	"context"
	"desafio-api/internal/domain"
	"desafio-api/internal/ports/repository"
)
type ItemService struct {
	repo repository.ItemRepository
}
func NewItemService(repo repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}
func (s *ItemService) Create(ctx context.Context, item *domain.Item) error {
	if err := item.Validate(); err != nil {
		return err
	}
	exists, err := s.repo.ExistsByCode(ctx, item.Code, 0)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrDuplicateCode
	}
	if item.Stock > 0 {
		item.Status = "ACTIVE"
	} else {
		item.Status = "INACTIVE"
	}
	if userID, ok := ctx.Value("userID").(int); ok {
		item.CreatedBy = userID
		item.UpdatedBy = userID
	}
	return s.repo.Save(ctx, item)
}
func (s *ItemService) Update(ctx context.Context, id int64, item *domain.Item) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	existing.Code = item.Code
	existing.Title = item.Title
	existing.Description = item.Description
	existing.Price = item.Price
	existing.Stock = item.Stock
	if existing.Stock > 0 {
		existing.Status = "ACTIVE"
	} else {
		existing.Status = "INACTIVE"
	}
	if userID, ok := ctx.Value("userID").(int); ok {
		existing.UpdatedBy = userID
	}
	if err := existing.Validate(); err != nil {
		return err
	}
	exists, err := s.repo.ExistsByCode(ctx, existing.Code, id)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrDuplicateCode
	}
	err = s.repo.Update(ctx, existing)
	if err != nil {
		return err
	}
	*item = *existing
	return nil
}
func (s *ItemService) GetByID(ctx context.Context, id int64) (*domain.Item, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *ItemService) List(ctx context.Context, status string, page, limit int) ([]*domain.Item, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 20 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.repo.FindAll(ctx, status, limit, offset)
}
func (s *ItemService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
