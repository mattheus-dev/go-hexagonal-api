package main
import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"desafio-api/internal/adapters/database"
	httpHandler "desafio-api/internal/adapters/http"
	"desafio-api/internal/adapters/repository"
	"desafio-api/internal/application/service"
	repoPort "desafio-api/internal/ports/repository"
	"desafio-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	cfg := loadConfig()
	log.Printf("DB Config: Host=%s, Port=%s, User=%s, DBName=%s", 
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName)
	var itemRepo repoPort.ItemRepository
	var userRepo service.UserRepository
	db, err := database.NewDB(database.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Password:     cfg.DBPassword,
		DBName:       cfg.DBName,
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 25,
		MaxIdleTime:  15 * time.Minute,
	})
	if err != nil {
		log.Printf(" Erro ao conectar ao banco de dados: %v", err)
		log.Println(" Usando repositórios simulados para demonstração")
		itemRepo = repository.NewMockItemRepository()
		userRepo = repository.NewMockUserRepository()
	} else {
		log.Println(" Conexão com o banco de dados estabelecida")
		itemRepo = repository.NewItemRepository(db)
		userRepo = repository.NewUserRepository(db)
		defer db.Close()
		if err := runMigrations(db); err != nil {
			log.Printf(" Erro ao executar migrações: %v", err)
		}
	}
	itemService := service.NewItemService(itemRepo)
	userService := service.NewUserService(userRepo)
	itemHandler := httpHandler.NewItemHandler(itemService)
	authHandler := httpHandler.NewAuthHandler(userService)
	router := setupRouter(itemHandler, authHandler, userService, db, cfg.DBName)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		log.Printf("Server is running on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
type Config struct {
	Port       int
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}
func loadConfig() Config {
	return Config{
		Port:       8080,
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "mercadolibre_challenge"),
		JWTSecret:  getEnv("JWT_SECRET", "your-default-jwt-secret-for-development"),
	}
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
func setupRouter(itemHandler *httpHandler.ItemHandler, authHandler *httpHandler.AuthHandler, userService *service.UserService, db *sqlx.DB, dbName string) *gin.Engine {
	if gin.Mode() == gin.DebugMode {
		log.Println("Running in DEBUG mode")
	}
	router := gin.New()
	router.Use(httpHandler.LoggingMiddleware()) 
	router.Use(httpHandler.ErrorMiddleware())   
	router.Use(gin.Recovery())                  
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"time": time.Now().Format(time.RFC3339),
		})
	})
	router.GET("/debug/database", func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(500, gin.H{
				"status": "erro",
				"error": err.Error(),
			})
			return
		}
		var userCount int
		err = db.Get(&userCount, "SELECT COUNT(*) FROM users")
		var itemCount int
		err2 := db.Get(&itemCount, "SELECT COUNT(*) FROM items")
		var dbErrorMsg, itemErrorMsg interface{}
		if err != nil {
			dbErrorMsg = err.Error()
		}
		if err2 != nil {
			itemErrorMsg = err2.Error()
		}
		c.JSON(200, gin.H{
			"status": "conectado",
			"database": dbName,
			"user_count": userCount,
			"item_count": itemCount,
			"db_error": dbErrorMsg,
			"item_error": itemErrorMsg,
		})
	})
	router.POST("/debug/register-test", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		user := &domain.User{
			Username: req.Username,
			Password: req.Password,
		}
		log.Printf("[DEBUG] Test Register: Starting validation")
		if err := user.Validate(); err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Validation failed: %v", err)})
			return
		}
		log.Printf("[DEBUG] Test Register: Starting password hash")
		if err := user.HashPassword(); err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Hash failed: %v", err)})
			return
		}
		log.Printf("[DEBUG] Test Register: Starting repository operation")
		err := userService.Register(c.Request.Context(), user)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Repository error: %v", err)})
			return
		}
		c.JSON(201, gin.H{
			"status": "success",
			"user_id": user.ID,
			"username": user.Username,
		})
	})
	router.GET("/debug/user-test", func(c *gin.Context) {
		user := &domain.User{
			Username: "debuguser",
			Password: "password123",
		}
		if err := user.HashPassword(); err != nil {
			log.Printf("Erro ao criptografar senha: %v", err)
			c.JSON(500, gin.H{"error": "Falha ao criptografar senha"})
			return
		}
		repoType := "unknown"
		switch r := userService.GetRepository().(type) {
		case *repository.UserRepository:
			repoType = "MySQL"
		case *repository.MockUserRepository:
			repoType = "Mock"
		default:
			repoType = fmt.Sprintf("%T", r)
		}
		log.Printf("Tentando salvar usuário no repositório tipo: %s", repoType)
		err := userService.Register(c.Request.Context(), user)
		if err != nil {
			log.Printf("Erro ao registrar usuário: %v", err)
			c.JSON(500, gin.H{
				"error": fmt.Sprintf("Falha ao registrar usuário: %v", err),
				"repository_type": repoType,
			})
			return
		}
		c.JSON(200, gin.H{
			"status": "Usuário de teste criado com sucesso",
			"user_id": user.ID,
			"repository_type": repoType,
		})
	})
	router.GET("/debug/jwt-test", func(c *gin.Context) {
		user := &domain.User{
			ID:       1,
			Username: "test",
		}
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "your-super-secret-jwt-key"
		}
		claims := &domain.JWTClaims{
			UserID:   1,
			Username: "test",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			log.Printf("Erro ao assinar token: %v", err)
			c.JSON(500, gin.H{
				"error": "Falha ao gerar token: " + err.Error(),
				"secret_length": len(secret),
			})
			return
		}
		c.JSON(200, gin.H{
			"token": tokenString,
			"user":  user,
			"secret_used": secret,
		})
	})
	router.GET("/debug/insert-user", func(c *gin.Context) {
		log.Println("Iniciando teste de inserção direta de usuário")
		username := "testuser_" + strconv.FormatInt(time.Now().Unix(), 10)
		password := "password123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Falha ao gerar hash: " + err.Error()})
			return
		}
		query := "INSERT INTO users (username, password) VALUES (?, ?)"
		result, err := db.ExecContext(c.Request.Context(), query, username, string(hashedPassword))
		if err != nil {
			log.Printf("[ERROR] Falha ao inserir usuário: %v", err)
			c.JSON(500, gin.H{"error": "Falha ao inserir usuário: " + err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		c.JSON(200, gin.H{
			"message": "Usuário inserido com sucesso via SQL direto",
			"user_id": id,
			"username": username,
		})
	})
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)
	v1 := router.Group("/api/v1")
	v1.Use(httpHandler.AuthMiddleware(userService))
	{
		items := v1.Group("/items")
		{
			items.POST("", itemHandler.Create)
			items.GET("", itemHandler.List)
			items.GET("/:id", itemHandler.GetByID)
			items.PUT("/:id", itemHandler.Update)
			items.DELETE("/:id", itemHandler.Delete)
		}
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	return router
}
func runMigrations(db *sqlx.DB) error {
	var tableItems string
	err := db.Get(&tableItems, `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = DATABASE() 
        AND table_name = 'items'
        LIMIT 1
    `)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Warning: items table does not exist. Please run migrations first.")
			log.Println("You can run the SQL migrations from the migrations/001_create_items_table.mysql.sql file.")
			return nil
		}
		return fmt.Errorf("failed to check if items table exists: %w", err)
	}
	log.Printf("Table 'items' exists: %s\n", tableItems)
	var tableUsers string
	err = db.Get(&tableUsers, `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = DATABASE() 
        AND table_name = 'users'
        LIMIT 1
    `)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if users table exists: %w", err)
	}
	if err == sql.ErrNoRows {
		log.Println("Warning: users table does not exist. Please run migrations first.")
		log.Println("You can run the SQL migrations from the migrations/002_create_users_table.mysql.sql file.")
	} else {
		log.Printf("Table 'users' exists: %s\n", tableUsers)
	}
	return nil
}
