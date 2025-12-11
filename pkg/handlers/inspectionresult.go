package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// ХЕНДЛЕР
// ============================================================================

type InspectionResultHandler struct {
	Service *service.InspectionResultService
}

func NewInspectionResultHandler(s *service.InspectionResultService) *InspectionResultHandler {
	return &InspectionResultHandler{Service: s}
}

// ============================================================================
// HTTP-ОБРАБОТЧИКИ
// ============================================================================

// CreateOrUpdateResult godoc
// @Summary      Создать/обновить результат осмотра
// @Description  Создание или обновление результата проверки элемента чек-листа
// @Tags         Инспектор
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Param        request body models.CreateInspectionResultRequest true "Результат осмотра элемента"
// @Success      201 {object} models.InspectionResultResponse "Результат сохранен"
// @Failure      400 {object} map[string]string "Неверный запрос или задание не в работе"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks/{id}/results [post]
func (h *InspectionResultHandler) CreateOrUpdateResult(c *gin.Context) {
	taskID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.CreateInspectionResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.CreateOrUpdateResult(c.Request.Context(), taskID, req)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrTaskNotInProgress) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task is not in progress (cannot add results)"})
			return
		}
		if errors.Is(err, service.ErrChecklistElementInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Checklist element does not belong to task's checklist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save result"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetTaskResults godoc
// @Summary      Получить результаты осмотра
// @Description  Возвращает все результаты осмотра для конкретного задания
// @Tags         Инспектор
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      200 {array} models.InspectionResultResponse "Список результатов осмотра"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks/{id}/results [get]
func (h *InspectionResultHandler) GetTaskResults(c *gin.Context) {
	taskID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	resp, err := h.Service.GetTaskResults(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve results"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteResult godoc
// @Summary      Удалить результат осмотра
// @Description  Удаление результата осмотра конкретного элемента
// @Tags         Инспектор
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Param        element_id path int true "ID элемента чек-листа"
// @Success      204 "Результат успешно удален"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Результат не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks/{id}/results/{element_id} [delete]
func (h *InspectionResultHandler) DeleteResult(c *gin.Context) {
	taskID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	elementID, err := strconv.Atoi(c.Param("element_id"))
	if err != nil || elementID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
		return
	}

	err = h.Service.DeleteResult(c.Request.Context(), taskID, elementID)
	if err != nil {
		if errors.Is(err, service.ErrResultNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete result"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
