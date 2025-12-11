// handlers/inspectionact.go

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

type InspectionActHandler struct {
	Service *service.InspectionActService
}

func NewInspectionActHandler(s *service.InspectionActService) *InspectionActHandler {
	return &InspectionActHandler{Service: s}
}

// DownloadAct godoc
// @Summary      Скачать акт осмотра
// @Description  Скачивание PDF-акта осмотра здания
// @Tags         Инспектор
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      200 {file} file "PDF файл акта осмотра"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Акт осмотра не найден"
// @Failure      500 {object} map[string]string "Ошибка генерации акта"
// @Router       /inspector/tasks/{id}/act [get]
func (h *InspectionActHandler) DownloadAct(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	pdfData, filename, err := h.Service.GeneratePDFForAct(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, service.ErrActNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Inspection act not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate inspection act"})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdfData)
}
