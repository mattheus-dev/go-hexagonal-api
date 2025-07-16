package repository
import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"desafio-api/internal/domain"
	repoPort "desafio-api/internal/ports/repository"
	"github.com/jmoiron/sqlx"
)
var _ repoPort.ItemRepository = (*itemRepository)(nil)
type itemRepository struct {
	db *sqlx.DB
}
func NewItemRepository(db *sqlx.DB) *itemRepository {
	return &itemRepository{db: db}
}
func (r *itemRepository) Save(ctx context.Context, item *domain.Item) error {
	query := `
        INSERT INTO items (code, title, description, price, stock, status, created_at, updated_at, created_by, updated_by)
        VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), ?, ?)`
    result, err := r.db.ExecContext(
        ctx,
        query,
        item.Code,
        item.Title,
        item.Description,
        item.Price,
        item.Stock,
        item.Status,
        item.CreatedBy,
        item.UpdatedBy,
    )
    if err != nil {
        return err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    item.ID = id
    item.CreatedAt = time.Now()
    item.UpdatedAt = time.Now()
    return nil
}
func (r *itemRepository) Update(ctx context.Context, item *domain.Item) error {
	query := `
        UPDATE items 
        SET code = ?, title = ?, description = ?, price = ?, stock = ?, status = ?, updated_at = NOW(), updated_by = ?
        WHERE id = ?`
    _, err := r.db.ExecContext(
        ctx,
        query,
        item.Code,
        item.Title,
        item.Description,
        item.Price,
        item.Stock,
        item.Status,
        item.UpdatedBy,
        item.ID,
    )
    if err != nil {
        return err
    }
    item.UpdatedAt = time.Now()
    return nil
}
func (r *itemRepository) FindByID(ctx context.Context, id int64) (*domain.Item, error) {
	var item domain.Item
	query := "SELECT * FROM items WHERE id = ?"
    err := r.db.GetContext(ctx, &item, query, id)
    if err == sql.ErrNoRows {
        return nil, domain.ErrItemNotFound
    }
    return &item, err
}
func (r *itemRepository) FindAll(ctx context.Context, status string, limit, offset int) ([]*domain.Item, int, error) {
	var items []*domain.Item
	var count int
    whereClause := ""
    var args []interface{}
    if status != "" {
        whereClause = " WHERE status = ?"
        args = append(args, status)
    }
    countQuery := "SELECT COUNT(*) FROM items" + whereClause
    err := r.db.GetContext(ctx, &count, countQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count items: %w", err)
    }
    if count == 0 {
        return []*domain.Item{}, 0, nil
    }
    query := "SELECT * FROM items" + whereClause + " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
    args = append(args, limit, offset)
    if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
        return nil, 0, fmt.Errorf("failed to fetch items: %w", err)
    }
    return items, count, nil
}
func (r *itemRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM items WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return domain.ErrItemNotFound
    }
    return nil
}
func (r *itemRepository) ExistsByCode(ctx context.Context, code string, excludeID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM items WHERE code = ? AND id != ?)"
    err := r.db.GetContext(ctx, &exists, query, code, excludeID)
    return exists, err
}
