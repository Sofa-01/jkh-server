package handlers

import (
    "errors"
    "net/http"
    "strings"

    "jkh/pkg/models"
    "jkh/pkg/service"

    "github.com/gin-gonic/gin"
)

// ============================================================================
// ХЕНДЛЕР
// ============================================================================

// ElementCatalogHandler — слой HTTP (Transport Layer).
// Обрабатывает входящие HTTP-запросы, валидирует данные и вызывает Service-слой.
type ElementCatalogHandler struct {
    Service *service.ElementCatalogService // Сервис для бизнес-логики
}

// NewElementCatalogHandler — конструктор хендлера.
func NewElementCatalogHandler(s *service.ElementCatalogService) *ElementCatalogHandler {
    return &ElementCatalogHandler{Service: s}
}

// ============================================================================
// HTTP-ОБРАБОТЧИКИ (соответствуют REST API)
// ============================================================================

// CreateElement godoc
// @Summary      Создать элемент справочника
// @Description  Создание нового элемента для использования в чек-листах (например: Фундамент, Кровля)
// @Tags         Справочник элементов
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateElementCatalogRequest true "Данные элемента"
// @Success      201 {object} models.ElementCatalogResponse "Элемент успешно создан"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Элемент с таким именем уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/elements [post]
func (h *ElementCatalogHandler) CreateElement(c *gin.Context) {
    var req models.CreateElementCatalogRequest
    
    // 1. Парсинг и валидация JSON из тела запроса
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    // 2. Вызов сервисного слоя для создания элемента
    resp, err := h.Service.CreateElement(c.Request.Context(), req)
    if err != nil {
        // Обработка специфичных ошибок бизнес-логики
        if errors.Is(err, service.ErrElementConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "Element name already exists"})
            return
        }
        // Общая ошибка БД
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create element"})
        return
    }

    // 3. Возврат успешного ответа с кодом 201 (Created)
    c.JSON(http.StatusCreated, resp)
}

// ListElements godoc
// @Summary      Получить список элементов справочника
// @Description  Возвращает список всех элементов для чек-листов
// @Tags         Справочник элементов
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.ElementCatalogResponse "Список элементов"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/elements [get]
func (h *ElementCatalogHandler) ListElements(c *gin.Context) {
    // Вызов сервиса для получения списка
    resp, err := h.Service.ListElements(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve element list"})
        return
    }
    
    // Возврат массива элементов (даже если пустой)
    c.JSON(http.StatusOK, resp)
}

// GetElement godoc
// @Summary      Получить элемент по ID
// @Description  Возвращает информацию о конкретном элементе справочника
// @Tags         Справочник элементов
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID элемента"
// @Success      200 {object} models.ElementCatalogResponse "Данные элемента"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Элемент не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/elements/{id} [get]
func (h *ElementCatalogHandler) GetElement(c *gin.Context) {
    // 1. Извлечение и валидация ID из URL-параметра
    // parseID — вспомогательная функция (определена в handlers/utils.go или building.go)
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
        return
    }

    // 2. Вызов сервиса для поиска элемента
    resp, err := h.Service.RetrieveElement(c.Request.Context(), id)
    if err != nil {
        // Элемент не найден
        if errors.Is(err, service.ErrElementNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Element not found"})
            return
        }
        // Ошибка БД
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve element"})
        return
    }

    // 3. Возврат найденного элемента
    c.JSON(http.StatusOK, resp)
}

// UpdateElement godoc
// @Summary      Обновить элемент
// @Description  Обновление данных элемента справочника
// @Tags         Справочник элементов
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID элемента"
// @Param        request body models.CreateElementCatalogRequest true "Данные для обновления"
// @Success      200 {object} models.ElementCatalogResponse "Обновленные данные элемента"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Элемент не найден"
// @Failure      409 {object} map[string]string "Имя элемента уже занято"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/elements/{id} [put]
func (h *ElementCatalogHandler) UpdateElement(c *gin.Context) {
    // 1. Валидация ID
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
        return
    }

    // 2. Парсинг тела запроса
    var req models.CreateElementCatalogRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    // 3. Вызов сервиса для обновления
    resp, err := h.Service.UpdateElement(c.Request.Context(), id, req)
    if err != nil {
        // Обработка ошибок
        if errors.Is(err, service.ErrElementNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Element not found"})
            return
        }
        if errors.Is(err, service.ErrElementConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "Element name already exists"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update element"})
        return
    }

    // 4. Возврат обновленного элемента
    c.JSON(http.StatusOK, resp)
}

// DeleteElement godoc
// @Summary      Удалить элемент
// @Description  Удаление элемента из справочника
// @Tags         Справочник элементов
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID элемента"
// @Success      204 "Элемент успешно удален"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Элемент не найден"
// @Failure      409 {object} map[string]string "Элемент используется в чек-листах"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/elements/{id} [delete]
func (h *ElementCatalogHandler) DeleteElement(c *gin.Context) {
    // 1. Валидация ID
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
        return
    }

    // 2. Вызов сервиса для удаления
    err = h.Service.DeleteElement(c.Request.Context(), id)
    if err != nil {
        // Обработка ошибок
        if errors.Is(err, service.ErrElementNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Element not found"})
            return
        }
        // FK constraint violation — элемент используется в чек-листах
        if strings.Contains(err.Error(), "active dependencies") {
            c.JSON(http.StatusConflict, gin.H{"error": "Element has active dependencies (used in checklists)"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete element"})
        return
    }

    // 3. Успешное удаление (без тела ответа)
    c.JSON(http.StatusNoContent, nil)
}
