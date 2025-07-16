package main
import (
	"context"
	"fmt"
	"log"
	"os"
	"desafio-api/internal/adapters/database"
	"desafio-api/internal/adapters/repository"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	fmt.Println("=== Diagnóstico Direto da API ===")
	_ = godotenv.Load()
	db, err := database.NewDB(database.Config{
		Host:     getEnv("DB_HOST", "db"),  
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "desafio_db"),
	})
	if err != nil {
		log.Fatalf("ERRO: Falha na conexão com o banco de dados: %v", err)
	}
	fmt.Println("✓ Conexão com banco de dados estabelecida")
	userRepo := repository.NewUserRepository(db)
	fmt.Println("✓ Repositório de usuários criado")
	fmt.Println("\n=== Teste 1: Buscar usuário existente ===")
	user, err := userRepo.FindByUsername(context.Background(), "direct_test_user")
	if err != nil {
		fmt.Printf("✗ ERRO: Não foi possível encontrar o usuário 'direct_test_user': %v\n", err)
	} else {
		fmt.Printf("✓ Usuário encontrado: ID=%d, Username=%s\n", user.ID, user.Username)
		fmt.Println("\n=== Teste 2: Validar senha do usuário ===")
		password := "password123"
		isValid := comparePassword(user.Password, password)
		if isValid {
			fmt.Printf("✓ Senha válida para o usuário '%s'\n", user.Username)
		} else {
			fmt.Printf("✗ ERRO: Senha inválida para o usuário '%s'\n", user.Username)
		}
	}
	fmt.Println("\n=== Teste 3: Criar novo usuário ===")
	newUser := &domain.User{
		Username: fmt.Sprintf("testuser_%d", os.Getpid()),
		Password: "password123",
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("✗ ERRO: Falha ao gerar hash da senha: %v\n", err)
	} else {
		newUser.Password = string(hashedPassword)
		fmt.Printf("✓ Hash da senha gerado: %s\n", newUser.Password[:20]+"...")
		err = userRepo.Create(context.Background(), newUser)
		if err != nil {
			fmt.Printf("✗ ERRO: Falha ao criar novo usuário: %v\n", err)
		} else {
			fmt.Printf("✓ Novo usuário criado: ID=%d, Username=%s\n", newUser.ID, newUser.Username)
		}
	}
	fmt.Println("\n=== Teste 4: Testar serviço de usuário ===")
	userService := service.NewUserService(userRepo)
	fmt.Println("✓ Serviço de usuário criado")
	fmt.Println("\n=== Teste 5: Registrar usuário via serviço ===")
	serviceUser := &domain.User{
		Username: fmt.Sprintf("serviceuser_%d", os.Getpid()),
		Password: "password123", 
	}
	err = userService.Register(context.Background(), serviceUser)
	if err != nil {
		fmt.Printf("✗ ERRO: Falha ao registrar usuário via serviço: %v\n", err)
	} else {
		fmt.Printf("✓ Usuário registrado via serviço: ID=%d, Username=%s\n", serviceUser.ID, serviceUser.Username)
	}
	fmt.Println("\n=== Teste 6: Login via serviço ===")
	token, err := userService.Login(context.Background(), serviceUser.Username, "password123")
	if err != nil {
		fmt.Printf("✗ ERRO: Falha ao fazer login: %v\n", err)
	} else {
		fmt.Printf("✓ Login bem-sucedido, token gerado: %s...\n", token[:20])
	}
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
func comparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
