package http
import (
	"log"
	"net/http"
	"strconv"
	"time"
	"desafio-api/internal/application/service"
	"desafio-api/internal/domain"
	"github.com/gin-gonic/gin"
)
type ItemHandler struct {
	itemService service.ItemServiceInterface
}
func NewItemHandler(itemService service.ItemServiceInterface) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}
type CreateRequest struct {
	Code        string `json:"code" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Stock       int    `json:"stock" binding:"gte=0"`
}
type UpdateRequest struct {
	Code        string `json:"code" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Stock       int    `json:"stock" binding:"gte=0"`
}
type ItemResponse struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CreatedBy   int    `json:"created_by"`
	UpdatedBy   int    `json:"updated_by"`
}
type ListResponse struct {
	TotalPages int             `json:"totalPages"`
	Data       []*ItemResponse `json:"data"`
}
func (h *ItemHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "Falha ao identificar o usuário autenticado")
		return
	}
	item := &domain.Item{
		Code:        req.Code,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      "INACTIVE", 
		CreatedBy:   userID.(int), 
		UpdatedBy:   userID.(int), 
	}
	if err := h.itemService.Create(c.Request.Context(), item); err != nil {
		switch {
		case err == domain.ErrDuplicateCode:
			RespondWithError(c, http.StatusConflict, "Já existe um item com este código")
		case err == domain.ErrCodeRequired || err == domain.ErrTitleRequired || 
			err == domain.ErrDescriptionRequired || err == domain.ErrInvalidPrice || 
			err == domain.ErrInvalidStock:
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			RespondWithError(c, http.StatusInternalServerError, "Falha ao criar o item")
		}
		return
	}
	c.JSON(http.StatusCreated, toItemResponse(item))
}
func (h *ItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "ID de item inválido")
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "Falha ao identificar o usuário autenticado")
		return
	}
	existingItem, err := h.itemService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrItemNotFound:
			RespondWithError(c, http.StatusNotFound, "Item não encontrado")
		default:
			log.Printf("Erro ao buscar item %d para atualização: %v", id, err)
			RespondWithError(c, http.StatusInternalServerError, "Falha ao recuperar o item para atualização")
		}
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}
	existingItem.Code = req.Code
	existingItem.Title = req.Title
	existingItem.Description = req.Description
	existingItem.Price = req.Price
	existingItem.Stock = req.Stock
	existingItem.UpdatedBy = userID.(int) 
	if err := h.itemService.Update(c.Request.Context(), id, existingItem); err != nil {
		switch {
		case err == domain.ErrItemNotFound:
			RespondWithError(c, http.StatusNotFound, "Item não encontrado")
		case err == domain.ErrDuplicateCode:
			RespondWithError(c, http.StatusConflict, "Já existe um item com este código")
		case err == domain.ErrCodeRequired || err == domain.ErrTitleRequired || 
			err == domain.ErrDescriptionRequired || err == domain.ErrInvalidPrice || 
			err == domain.ErrInvalidStock:
			RespondWithError(c, http.StatusBadRequest, err.Error())
		default:
			log.Printf("Erro ao atualizar item %d: %v", id, err)
			RespondWithError(c, http.StatusInternalServerError, "Falha ao atualizar o item")
		}
		return
	}
	c.JSON(http.StatusOK, toItemResponse(existingItem))
}
func (h *ItemHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "ID de item inválido")
		return
	}
	item, err := h.itemService.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrItemNotFound:
			RespondWithError(c, http.StatusNotFound, "Item não encontrado")
		default:
			log.Printf("Erro ao buscar item %d: %v", id, err)
			RespondWithError(c, http.StatusInternalServerError, "Falha ao buscar o item")
		}
		return
	}
	c.JSON(http.StatusOK, toItemResponse(item))
}
func (h *ItemHandler) List(c *gin.Context) {
	status := c.Query("status")
	if status != "" && status != "ACTIVE" && status != "INACTIVE" {
		RespondWithError(c, http.StatusBadRequest, "Status inválido. Use 'ACTIVE' ou 'INACTIVE'")
		return
	}
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		RespondWithError(c, http.StatusBadRequest, "O parâmetro 'limit' deve ser um número entre 1 e 100")
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		RespondWithError(c, http.StatusBadRequest, "O parâmetro 'page' deve ser um número maior que zero")
		return
	}
	offset := (page - 1) * limit
	items, total, err := h.itemService.List(c.Request.Context(), status, limit, offset)
	if err != nil {
		log.Printf("Erro ao listar itens: %v", err)
		RespondWithError(c, http.StatusInternalServerError, "Falha ao recuperar a lista de itens")
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	response := ListResponse{
		TotalPages: totalPages,
		Data:       make([]*ItemResponse, 0, len(items)),
	}
	for _, item := range items {
		response.Data = append(response.Data, toItemResponse(item))
	}
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Page", strconv.Itoa(page))
	c.Header("X-Per-Page", strconv.Itoa(limit))
	c.Header("X-Total-Pages", strconv.Itoa(totalPages))
	c.JSON(http.StatusOK, response)
}
func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "ID de item inválido")
		return
	}
	if err := h.itemService.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case domain.ErrItemNotFound:
			RespondWithError(c, http.StatusNotFound, "Item não encontrado")
		default:
			log.Printf("Erro ao remover item %d: %v", id, err)
			RespondWithError(c, http.StatusInternalServerError, "Falha ao remover o item")
		}
		return
	}
	c.Status(http.StatusNoContent)
}
func toItemResponse(item *domain.Item) *ItemResponse {
	if item == nil {
		return nil
	}
	response := &ItemResponse{
		ID:          item.ID,
		Code:        item.Code,
		Title:       item.Title,
		Description: item.Description,
		Price:       item.Price,
		Stock:       item.Stock,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
	}
	if !item.CreatedAt.IsZero() {
		response.CreatedAt = item.CreatedAt.Format(time.RFC3339)
	}
	if !item.UpdatedAt.IsZero() {
		response.UpdatedAt = item.UpdatedAt.Format(time.RFC3339)
	}
	return response
}
