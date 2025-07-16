package domain
import "errors"
var (
    ErrItemNotFound      = errors.New("item not found")
    ErrCodeRequired      = errors.New("code is required")
    ErrTitleRequired     = errors.New("title is required")
    ErrDescriptionRequired = errors.New("description is required")
    ErrInvalidPrice      = errors.New("price must be greater than zero")
    ErrInvalidStock      = errors.New("stock cannot be negative")
    ErrDuplicateCode     = errors.New("item with this code already exists")
    ErrUserNotFound      = errors.New("user not found")
    ErrUsernameRequired  = errors.New("username is required")
    ErrPasswordTooShort  = errors.New("password must be at least 6 characters")
    ErrDuplicateUsername = errors.New("user with this username already exists")
    ErrInvalidCredentials = errors.New("invalid username or password")
    ErrInvalidToken      = errors.New("invalid or expired token")
)
