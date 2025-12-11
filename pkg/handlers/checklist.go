package handlers

import (
    "errors"
    "net/http"
    "strconv"
    "strings"

    "jkh/pkg/models"
    "jkh/pkg/service"

    "github.com/gin-gonic/gin"
)

// ============================================================================
// ХЕНДЛЕР
// ============================================================================

// ChecklistHandler — слой HTTP для работы с чек-листами.
type ChecklistHandler struct {
    Service *service.ChecklistService
}

func NewChecklistHandler(s *service.ChecklistService) *ChecklistHandler {
    return &ChecklistHandler{Service: s}
}

// ============================================================================
// CRUD ДЛЯ CHECKLIST
// ============================================================================

// CreateChecklist godoc
// @Summary      Создать чек-лист
// @Description  Создание нового чек-листа для осмотра зданий
// @Tags         Чек-листы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateChecklistRequest true "Данные чек-листа"
// @Success      201 {object} models.ChecklistResponse "Чек-лист успешно создан"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Чек-лист с таким названием уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists [post]
func (h *ChecklistHandler) CreateChecklist(c *gin.Context) {
    var req models.CreateChecklistRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    resp, err := h.Service.CreateChecklist(c.Request.Context(), req)
    if err != nil {
        if errors.Is(err, service.ErrChecklistConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "Checklist title already exists"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checklist"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}

// ListChecklists godoc
// @Summary      Получить список чек-листов
// @Description  Возвращает список всех чек-листов (без детализации элементов)
// @Tags         Чек-листы
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.ChecklistResponse "Список чек-листов"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists [get]
func (h *ChecklistHandler) ListChecklists(c *gin.Context) {
    resp, err := h.Service.ListChecklists(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checklist list"})
        return
    }
    c.JSON(http.StatusOK, resp)
}

// GetChecklist godoc
// @Summary      Получить чек-лист по ID
// @Description  Возвращает детальную информацию о чек-листе включая список элементов
// @Tags         Чек-листы
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Success      200 {object} models.ChecklistDetailResponse "Данные чек-листа с элементами"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Чек-лист не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id} [get]
func (h *ChecklistHandler) GetChecklist(c *gin.Context) {
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    resp, err := h.Service.RetrieveChecklist(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrChecklistNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checklist"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// UpdateChecklist godoc
// @Summary      Обновить чек-лист
// @Description  Обновление данных чек-листа (название, тип осмотра)
// @Tags         Чек-листы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Param        request body models.CreateChecklistRequest true "Данные для обновления"
// @Success      200 {object} models.ChecklistResponse "Обновленные данные чек-листа"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Чек-лист не найден"
// @Failure      409 {object} map[string]string "Название чек-листа уже занято"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id} [put]
func (h *ChecklistHandler) UpdateChecklist(c *gin.Context) {
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    var req models.CreateChecklistRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    resp, err := h.Service.UpdateChecklist(c.Request.Context(), id, req)
    if err != nil {
        if errors.Is(err, service.ErrChecklistNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
            return
        }
        if errors.Is(err, service.ErrChecklistConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "Checklist title already exists"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update checklist"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// DeleteChecklist godoc
// @Summary      Удалить чек-лист
// @Description  Удаление чек-листа из системы
// @Tags         Чек-листы
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Success      204 "Чек-лист успешно удален"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Чек-лист не найден"
// @Failure      409 {object} map[string]string "Чек-лист используется в заданиях"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id} [delete]
func (h *ChecklistHandler) DeleteChecklist(c *gin.Context) {
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    err = h.Service.DeleteChecklist(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, service.ErrChecklistNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
            return
        }
        if strings.Contains(err.Error(), "active dependencies") {
            c.JSON(http.StatusConflict, gin.H{"error": "Checklist has active dependencies (used in tasks)"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete checklist"})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}

// ============================================================================
// УПРАВЛЕНИЕ ЭЛЕМЕНТАМИ В ЧЕК-ЛИСТЕ
// ============================================================================

// AddElementToChecklist godoc
// @Summary      Добавить элемент в чек-лист
// @Description  Добавление элемента из справочника в чек-лист
// @Tags         Чек-листы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Param        request body models.AddElementToChecklistRequest true "ID элемента и порядок"
// @Success      201 {object} map[string]string "Элемент успешно добавлен"
// @Failure      400 {object} map[string]string "Неверный запрос или элемент не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Чек-лист не найден"
// @Failure      409 {object} map[string]string "Элемент уже добавлен в чек-лист"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id}/elements [post]
func (h *ChecklistHandler) AddElementToChecklist(c *gin.Context) {
    checklistID, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    var req models.AddElementToChecklistRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    err = h.Service.AddElementToChecklist(c.Request.Context(), checklistID, req)
    if err != nil {
        if errors.Is(err, service.ErrChecklistNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
            return
        }
        if errors.Is(err, service.ErrElementNotFound) {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Element not found in catalog"})
            return
        }
        if errors.Is(err, service.ErrElementAlreadyInChecklist) {
            c.JSON(http.StatusConflict, gin.H{"error": "Element already added to this checklist"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add element to checklist"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Element added to checklist successfully"})
}

// RemoveElementFromChecklist godoc
// @Summary      Удалить элемент из чек-листа
// @Description  Удаление элемента из конкретного чек-листа
// @Tags         Чек-листы
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Param        element_id path int true "ID элемента"
// @Success      204 "Элемент успешно удален из чек-листа"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Элемент не найден в чек-листе"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id}/elements/{element_id} [delete]
func (h *ChecklistHandler) RemoveElementFromChecklist(c *gin.Context) {
    checklistID, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    elementID, err := strconv.Atoi(c.Param("element_id"))
    if err != nil || elementID <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
        return
    }

    err = h.Service.RemoveElementFromChecklist(c.Request.Context(), checklistID, elementID)
    if err != nil {
        if errors.Is(err, service.ErrChecklistElementNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Element not found in this checklist"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove element from checklist"})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}

// UpdateElementOrder godoc
// @Summary      Изменить порядок элемента
// @Description  Изменение позиции элемента в чек-листе
// @Tags         Чек-листы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID чек-листа"
// @Param        element_id path int true "ID элемента"
// @Param        request body models.UpdateElementOrderRequest true "Новый порядковый номер"
// @Success      200 {object} map[string]string "Порядок успешно изменен"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Элемент не найден в чек-листе"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/checklists/{id}/elements/{element_id} [put]
func (h *ChecklistHandler) UpdateElementOrder(c *gin.Context) {
    checklistID, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
        return
    }

    elementID, err := strconv.Atoi(c.Param("element_id"))
    if err != nil || elementID <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
        return
    }

    var req models.UpdateElementOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    err = h.Service.UpdateElementOrder(c.Request.Context(), checklistID, elementID, req.OrderIndex)
    if err != nil {
        if errors.Is(err, service.ErrChecklistElementNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Element not found in this checklist"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update element order"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Element order updated successfully"})
}
