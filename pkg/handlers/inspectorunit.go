// pkg/handlers/inspectorunit.go

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

type InspectorUnitHandler struct {
	Service *service.InspectorUnitService
}

func NewInspectorUnitHandler(s *service.InspectorUnitService) *InspectorUnitHandler {
	return &InspectorUnitHandler{Service: s}
}

// AssignInspector godoc
// @Summary      Назначить инспектора на ЖЭУ
// @Description  Привязка инспектора к жилищно-эксплуатационной единице
// @Tags         Назначения инспекторов
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Param        request body models.AssignInspectorToUnitRequest true "ID инспектора"
// @Success      201 {object} models.InspectorAssignmentResponse "Инспектор успешно назначен"
// @Failure      400 {object} map[string]string "Неверный запрос или ЖЭУ/инспектор не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Инспектор уже назначен"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id}/inspectors [post]
func (h *InspectorUnitHandler) AssignInspector(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JKH unit ID"})
		return
	}

	var req models.AssignInspectorToUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = h.Service.AssignInspector(c.Request.Context(), id, req.InspectorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Inspector (user) not found"})
			return
		case errors.Is(err, service.ErrJkhUnitNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "JKH unit not found"})
			return
		case errors.Is(err, service.ErrInspectorAssignmentExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Inspector already assigned"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign inspector"})
			return
		}
	}

	resp := models.InspectorAssignmentResponse{
		JkhUnitID:   id,
		InspectorID: req.InspectorID,
		Message:     "assigned",
	}
	c.JSON(http.StatusCreated, resp)
}

// UnassignInspector godoc
// @Summary      Открепить инспектора от ЖЭУ
// @Description  Удаление привязки инспектора к жилищно-эксплуатационной единице
// @Tags         Назначения инспекторов
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Param        inspector_id path int true "ID инспектора"
// @Success      204 "Инспектор успешно откреплен"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Назначение не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id}/inspectors/{inspector_id} [delete]
func (h *InspectorUnitHandler) UnassignInspector(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JKH unit ID"})
		return
	}

	inspectorStr := c.Param("inspector_id")
	inspectorID, err := strconv.Atoi(inspectorStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspector ID"})
		return
	}

	err = h.Service.UnassignInspector(c.Request.Context(), id, inspectorID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInspectorAssignmentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unassign"})
			return
		}
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListInspectorsForUnit godoc
// @Summary      Получить инспекторов ЖЭУ
// @Description  Возвращает список инспекторов, привязанных к конкретному ЖЭУ
// @Tags         Назначения инспекторов
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID ЖЭУ"
// @Success      200 {array} models.UserResponse "Список инспекторов"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/jkhunits/{id}/inspectors [get]
func (h *InspectorUnitHandler) ListInspectorsForUnit(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JKH unit ID"})
		return
	}

	list, err := h.Service.ListInspectorsForUnit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list inspectors"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// ListUnitsForInspector godoc
// @Summary      Получить ЖЭУ инспектора
// @Description  Возвращает список ЖЭУ, к которым привязан конкретный инспектор
// @Tags         Назначения инспекторов
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID пользователя (инспектора)"
// @Success      200 {array} models.JkhUnitResponse "Список ЖЭУ"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users/{id}/jkhunits [get]
func (h *InspectorUnitHandler) ListUnitsForInspector(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	list, err := h.Service.ListUnitsForInspector(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list units"})
		return
	}
	c.JSON(http.StatusOK, list)
}
