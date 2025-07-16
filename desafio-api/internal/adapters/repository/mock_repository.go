package repository
import (
	"context"
	"sync"
	"time"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	repoPort "desafio-api/internal/ports/repository"
)
type MockItemRepository struct {
	items  map[int64]*domain.Item
	nextID int64
	mu     sync.RWMutex
}
func NewMockItemRepository() repoPort.ItemRepository {
	return &MockItemRepository{
		items:  make(map[int64]*domain.Item),
		nextID: 1,
	}
}
func (r *MockItemRepository) Save(ctx context.Context, item *domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existingItem := range r.items {
		if existingItem.Code == item.Code {
			return domain.ErrDuplicateCode
		}
	}
	item.ID = r.nextID
	r.nextID++
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	r.items[item.ID] = item
	return nil
}
func (r *MockItemRepository) Update(ctx context.Context, item *domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[item.ID]; !exists {
		return domain.ErrItemNotFound
	}
	for id, existingItem := range r.items {
		if existingItem.Code == item.Code && id != item.ID {
			return domain.ErrDuplicateCode
		}
	}
	item.UpdatedAt = time.Now()
	r.items[item.ID] = item
	return nil
}
func (r *MockItemRepository) FindByID(ctx context.Context, id int64) (*domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.items[id]
	if !exists {
		return nil, domain.ErrItemNotFound
	}
	return item, nil
}
func (r *MockItemRepository) FindAll(ctx context.Context, status string, limit, offset int) ([]*domain.Item, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var filteredItems []*domain.Item
	for _, item := range r.items {
		if status == "" || item.Status == status {
			filteredItems = append(filteredItems, item)
		}
	}
	total := len(filteredItems)
	if offset >= total {
		return []*domain.Item{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	paginatedItems := filteredItems[offset:end]
	return paginatedItems, total, nil
}
func (r *MockItemRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[id]; !exists {
		return domain.ErrItemNotFound
	}
	delete(r.items, id)
	return nil
}
func (r *MockItemRepository) ExistsByCode(ctx context.Context, code string, excludeID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, item := range r.items {
		if item.Code == code && id != excludeID {
			return true, nil
		}
	}
	return false, nil
}
type MockUserRepository struct {
	users  map[int]*domain.User
	nextID int
	mu     sync.RWMutex
}
var _ service.UserRepository = (*MockUserRepository)(nil)
func NewMockUserRepository() service.UserRepository {
	return &MockUserRepository{
		users:  make(map[int]*domain.User),
		nextID: 1,
	}
}
func (r *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existingUser := range r.users {
		if existingUser.Username == user.Username {
			return domain.ErrDuplicateUsername
		}
	}
	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}
func (r *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (r *MockUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}
