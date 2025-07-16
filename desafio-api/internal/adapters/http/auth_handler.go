package http
import (
	"log"
	"net/http"
	"time"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	"github.com/gin-gonic/gin"
)
type AuthHandler struct {
	userService service.UserServiceInterface
}
func NewAuthHandler(userService service.UserServiceInterface) *AuthHandler {
	return &AuthHandler{userService: userService}
}
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type LoginResponse struct {
	Token string `json:"token"`
}
func respondWithError(c *gin.Context, status int, message string) {
	log.Printf("[DEBUG] Respondendo com erro HTTP %d: %s", status, message)
	c.JSON(status, gin.H{
		"error": message,
		"status": status,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Register: Erro na validação de dados: %v", err)
		RespondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}
	log.Printf("[DEBUG] Register: Tentativa de registro para usuário: %s", req.Username)
	if len(req.Password) < 6 {
		log.Printf("[ERROR] Register: Senha muito curta para usuário %s", req.Username)
		RespondWithError(c, http.StatusBadRequest, "Senha deve ter pelo menos 6 caracteres")
		return
	}
	user := &domain.User{
		Username: req.Username,
		Password: req.Password,
	}
	err := h.userService.Register(c.Request.Context(), user)
	if err != nil {
		log.Printf("[ERROR] Register: Erro ao registrar usuário %s: %v", req.Username, err)
		switch err {
		case domain.ErrDuplicateUsername:
			RespondWithError(c, http.StatusConflict, "Usuário já existe")
		case domain.ErrUsernameRequired:
			RespondWithError(c, http.StatusBadRequest, "Nome de usuário é obrigatório")
		case domain.ErrPasswordTooShort:
			RespondWithError(c, http.StatusBadRequest, "Senha deve ter pelo menos 6 caracteres")
		default:
			log.Printf("[ERROR] Register: Erro interno no registro: %v", err)
			RespondWithError(c, http.StatusInternalServerError, "Erro interno ao registrar usuário")
		}
		return
	}
	log.Printf("[INFO] Register: Usuário %s registrado com sucesso (ID: %d)", user.Username, user.ID)
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username})
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Login: Erro na validação de dados: %v", err)
		RespondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}
	log.Printf("[DEBUG] Login: Tentativa de login para usuário: %s", req.Username)
	token, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("[ERROR] Login: Falha na autenticação para usuário %s: %v", req.Username, err)
		RespondWithError(c, http.StatusUnauthorized, "Credenciais inválidas")
		return
	}
	log.Printf("[INFO] Login: Usuário %s autenticado com sucesso", req.Username)
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
