//pkg/handlers/jkhunit.go

package handlers

import (
	"errors"
	"net/http"
	"strings"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

// JkhUnitHandler хранит сервис
type JkhUnitHandler struct {
	Service *service.JkhUnitService
}

func NewJkhUnitHandler(s *service.JkhUnitService) *JkhUnitHandler {
	return &JkhUnitHandler{Service: s}
}

// CreateJkhUnit godoc
// @Summary      Создать ЖЭУ
// @Description  Создание новой жилищно-эксплуатационной единицы
// @Tags         ЖЭУ
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateJkhUnitRequest true "Данные ЖЭУ"
// @Success      201 {object} models.JkhUnitResponse "ЖЭУ успешно создано"
// @Failure      400 {object} map[string]string "Неверный запрос или район не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "ЖЭУ с таким названием уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits [post]
func (h *JkhUnitHandler) CreateJkhUnit(c *gin.Context) {
	var req models.CreateJkhUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.Service.CreateJkhUnit(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrDistrictFKNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "District ID does not exist"})
			return
		}
		if errors.Is(err, service.ErrJkhUnitConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "JKH unit name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JKH unit"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListJkhUnits godoc
// @Summary      Получить список ЖЭУ
// @Description  Возвращает список всех жилищно-эксплуатационных единиц
// @Tags         ЖЭУ
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.JkhUnitResponse "Список ЖЭУ"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits [get]
func (h *JkhUnitHandler) ListJkhUnits(c *gin.Context) {
	resp, err := h.Service.ListJkhUnits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get list"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetJkhUnit godoc
// @Summary      Получить ЖЭУ по ID
// @Description  Возвращает информацию о конкретном ЖЭУ
// @Tags         ЖЭУ
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Success      200 {object} models.JkhUnitResponse "Данные ЖЭУ"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "ЖЭУ не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id} [get]
func (h *JkhUnitHandler) GetJkhUnit(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	resp, err := h.Service.RetrieveJkhUnit(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrJkhUnitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "JKH unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get JKH unit"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateJkhUnit godoc
// @Summary      Обновить ЖЭУ
// @Description  Обновление данных жилищно-эксплуатационной единицы
// @Tags         ЖЭУ
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Param        request body models.CreateJkhUnitRequest true "Данные для обновления"
// @Success      200 {object} models.JkhUnitResponse "Обновленные данные ЖЭУ"
// @Failure      400 {object} map[string]string "Неверный запрос или район не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "ЖЭУ не найдено"
// @Failure      409 {object} map[string]string "Название ЖЭУ уже занято"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id} [put]
func (h *JkhUnitHandler) UpdateJkhUnit(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.CreateJkhUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := h.Service.UpdateJkhUnit(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJkhUnitNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "JKH unit not found"})
		case errors.Is(err, service.ErrDistrictFKNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "District ID does not exist"})
		case errors.Is(err, service.ErrJkhUnitConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "JKH unit name already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteJkhUnit godoc
// @Summary      Удалить ЖЭУ
// @Description  Удаление жилищно-эксплуатационной единицы из системы
// @Tags         ЖЭУ
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Success      204 "ЖЭУ успешно удалено"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "ЖЭУ не найдено"
// @Failure      409 {object} map[string]string "У ЖЭУ есть активные зависимости"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id} [delete]
func (h *JkhUnitHandler) DeleteJkhUnit(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.Service.DeleteJkhUnit(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrJkhUnitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "JKH unit not found"})
			return
		}
		if strings.Contains(err.Error(), "dependencies") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete: JKH unit has dependencies"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
