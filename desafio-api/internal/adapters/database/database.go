package database
import (
    "context"
    "fmt"
    "time"
    _ "github.com/go-sql-driver/mysql"
    "github.com/jmoiron/sqlx"
)
type Config struct {
    Host         string
    Port         string
    User         string
    Password     string
    DBName       string
    SSLMode      string
    MaxOpenConns int
    MaxIdleConns int
    MaxIdleTime  time.Duration
}
func NewDB(cfg Config) (*sqlx.DB, error) {
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?parseTime=true",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
    )
    db, err := sqlx.Connect("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxIdleTime(cfg.MaxIdleTime)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    return db, nil
}
func WithTransaction(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
    tx, err := db.Beginx()
    if err != nil {
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()
    err = fn(tx)
    return err
}
