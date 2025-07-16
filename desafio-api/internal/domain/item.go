package domain
import "time"
type Item struct {
    ID          int64     `json:"id" db:"id"`
    Code        string    `json:"code" db:"code"`
    Title       string    `json:"title" db:"title"`
    Description string    `json:"description" db:"description"`
    Price       int64     `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    Status      string    `json:"status" db:"status"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    CreatedBy   int       `json:"created_by" db:"created_by"`
    UpdatedBy   int       `json:"updated_by" db:"updated_by"`
}
func (i *Item) Validate() error {
    if i.Code == "" {
        return ErrCodeRequired
    }
    if i.Title == "" {
        return ErrTitleRequired
    }
    if i.Description == "" {
        return ErrDescriptionRequired
    }
    if i.Price <= 0 {
        return ErrInvalidPrice
    }
    if i.Stock < 0 {
        return ErrInvalidStock
    }
    return nil
}
