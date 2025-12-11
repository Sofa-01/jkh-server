package handlers

import (
	"errors"
	"net/http"
	"strings"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

// BuildingHandler — слой HTTP (Transport).
type BuildingHandler struct {
	Service *service.BuildingService
}

func NewBuildingHandler(s *service.BuildingService) *BuildingHandler {
	return &BuildingHandler{Service: s}
}

// CreateBuilding godoc
// @Summary      Создать здание
// @Description  Создание нового здания для осмотра
// @Tags         Здания
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateBuildingRequest true "Данные здания"
// @Success      201 {object} models.BuildingResponse "Здание успешно создано"
// @Failure      400 {object} map[string]string "Неверный запрос или FK не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Адрес здания уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/buildings [post]
func (h *BuildingHandler) CreateBuilding(c *gin.Context) {
	var req models.CreateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.CreateBuilding(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrBuildingConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "Building address already exists"})
			return
		}
		if errors.Is(err, service.ErrFKNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid District, JKH Unit, or Inspector ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create building"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListBuildings godoc
// @Summary      Получить список зданий
// @Description  Возвращает список всех зданий в системе
// @Tags         Здания
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.BuildingResponse "Список зданий"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/buildings [get]
func (h *BuildingHandler) ListBuildings(c *gin.Context) {
	resp, err := h.Service.ListBuildings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve building list"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetBuilding godoc
// @Summary      Получить здание по ID
// @Description  Возвращает информацию о конкретном здании
// @Tags         Здания
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID здания"
// @Success      200 {object} models.BuildingResponse "Данные здания"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Здание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/buildings/{id} [get]
func (h *BuildingHandler) GetBuilding(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	resp, err := h.Service.RetrieveBuilding(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrBuildingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve building"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateBuilding godoc
// @Summary      Обновить здание
// @Description  Обновление данных здания
// @Tags         Здания
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID здания"
// @Param        request body models.CreateBuildingRequest true "Данные для обновления"
// @Success      200 {object} models.BuildingResponse "Обновленные данные здания"
// @Failure      400 {object} map[string]string "Неверный запрос или FK не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Здание не найдено"
// @Failure      409 {object} map[string]string "Адрес здания уже занят"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/buildings/{id} [put]
func (h *BuildingHandler) UpdateBuilding(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	var req models.CreateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.UpdateBuilding(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrBuildingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		if errors.Is(err, service.ErrBuildingConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "Building address already exists"})
			return
		}
		if errors.Is(err, service.ErrFKNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid District, JKH Unit, or Inspector ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update building"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteBuilding godoc
// @Summary      Удалить здание
// @Description  Удаление здания из системы
// @Tags         Здания
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID здания"
// @Success      204 "Здание успешно удалено"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Здание не найдено"
// @Failure      409 {object} map[string]string "У здания есть активные задания"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/buildings/{id} [delete]
func (h *BuildingHandler) DeleteBuilding(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	err = h.Service.DeleteBuilding(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrBuildingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		if strings.Contains(err.Error(), "active dependencies") {
			c.JSON(http.StatusConflict, gin.H{"error": "Building has active dependencies (tasks)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete building"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
