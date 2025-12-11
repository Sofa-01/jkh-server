package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"jkh/pkg/models"
	"jkh/pkg/service"
)

// DistrictHandler связывает HTTP-запросы с DistrictService
type DistrictHandler struct {
	Service *service.DistrictService
}

// Конструктор
func NewDistrictHandler(s *service.DistrictService) *DistrictHandler {
	return &DistrictHandler{Service: s}
}

// CreateDistrict godoc
// @Summary      Создать район
// @Description  Создание нового района города
// @Tags         Районы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateDistrictRequest true "Данные района"
// @Success      201 {object} models.DistrictResponse "Район успешно создан"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Район с таким названием уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/districts [post]
func (h *DistrictHandler) CreateDistrict(c *gin.Context) {
	var req models.CreateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.CreateDistrict(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrDistrictConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "District name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create district"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListDistricts godoc
// @Summary      Получить список районов
// @Description  Возвращает список всех районов города
// @Tags         Районы
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.DistrictResponse "Список районов"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/districts [get]
func (h *DistrictHandler) ListDistricts(c *gin.Context) {
	resp, err := h.Service.ListDistricts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve district list"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDistrict godoc
// @Summary      Получить район по ID
// @Description  Возвращает информацию о конкретном районе
// @Tags         Районы
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID района"
// @Success      200 {object} models.DistrictResponse "Данные района"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Район не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/districts/{id} [get]
func (h *DistrictHandler) GetDistrict(c *gin.Context) {
	id, err := parseID(c) // parseID из user.go
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid district ID"})
		return
	}

	resp, err := h.Service.RetrieveDistrict(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDistrictNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "District not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve district"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateDistrict godoc
// @Summary      Обновить район
// @Description  Обновление данных района
// @Tags         Районы
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID района"
// @Param        request body models.CreateDistrictRequest true "Данные для обновления"
// @Success      200 {object} models.DistrictResponse "Обновленные данные района"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Район не найден"
// @Failure      409 {object} map[string]string "Название района уже занято"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/districts/{id} [put]
func (h *DistrictHandler) UpdateDistrict(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid district ID"})
		return
	}

	var req models.CreateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.UpdateDistrict(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrDistrictNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "District not found"})
			return
		}
		if errors.Is(err, service.ErrDistrictConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "District name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update district"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteDistrict godoc
// @Summary      Удалить район
// @Description  Удаление района из системы
// @Tags         Районы
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID района"
// @Success      204 "Район успешно удален"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Район не найден"
// @Failure      409 {object} map[string]string "У района есть активные ЖЭУ или здания"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/districts/{id} [delete]
func (h *DistrictHandler) DeleteDistrict(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid district ID"})
		return
	}

	err = h.Service.DeleteDistrict(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDistrictNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "District not found"})
			return
		}
		if strings.Contains(err.Error(), "active dependencies") {
			c.JSON(http.StatusConflict, gin.H{"error": "District has active JKH units or buildings and cannot be deleted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete district"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
