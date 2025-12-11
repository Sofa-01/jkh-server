// pkg/handlers/analytics.go

package handlers

import (
	"fmt"
	"net/http"
	"time"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	Service *service.AnalyticsService
}

func NewAnalyticsHandler(s *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{Service: s}
}

// PreviewChart godoc
// @Summary      Предпросмотр графика
// @Description  Генерация графика в формате PNG для предпросмотра
// @Tags         Аналитика
// @Produce      image/png
// @Security     BearerAuth
// @Param        chart query string true "Тип графика" Enums(inspector_performance, status_distribution, failure_frequency)
// @Param        from query string true "Начало периода (YYYY-MM-DD)"
// @Param        to query string true "Конец периода (YYYY-MM-DD)"
// @Success      200 {file} file "PNG изображение графика"
// @Failure      400 {object} map[string]string "Неверные параметры"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Ошибка генерации графика"
// @Router       /tasks/analytics/preview [get]
func (h *AnalyticsHandler) PreviewChart(c *gin.Context) {
	chart := c.Query("chart")
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if chart == "" || fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing params"})
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
		return
	}

	var img []byte

	switch chart {
	case "inspector_performance":
		img, err = h.Service.GenerateInspectorPerformancePNG(c.Request.Context(), from, to)
	case "status_distribution":
		img, err = h.Service.GenerateStatusDistributionPNG(c.Request.Context(), from, to)
	case "failure_frequency":
		img, err = h.Service.GenerateFailureFrequencyPNG(c.Request.Context(), from, to)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported chart type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build chart: " + err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/png", img)
}

// GenerateReport godoc
// @Summary      Сгенерировать PDF отчёт
// @Description  Генерация аналитического PDF отчёта с графиками за указанный период
// @Tags         Аналитика
// @Accept       json
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        request body models.AnalyticsReportRequest true "Параметры отчёта"
// @Success      200 {file} file "PDF файл отчёта"
// @Failure      400 {object} map[string]string "Неверные параметры"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Ошибка генерации отчёта"
// @Router       /tasks/analytics/report [post]
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	var req models.AnalyticsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
		return
	}
	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
		return
	}

	charts := req.Charts
	if len(charts) == 0 {
		// По умолчанию генерируем все 3 графика
		charts = []string{"status_distribution", "failure_frequency", "inspector_performance"}
	}

	pdfBytes, filename, err := h.Service.GenerateReportPDF(c.Request.Context(), from, to, charts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
