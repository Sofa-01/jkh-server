// pkg/service/analytics.go

package service

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"sort"
	"time"

	"jkh/ent"
	"jkh/ent/inspectionresult"
	"jkh/ent/task"

	"github.com/jung-kurt/gofpdf"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// AnalyticsService отвечает за агрегации, построение графиков и генерацию PDF-отчётов
type AnalyticsService struct {
	Client *ent.Client
}

func NewAnalyticsService(client *ent.Client) *AnalyticsService {
	return &AnalyticsService{Client: client}
}

// GenerateInspectorPerformancePNG — простой пример: количество завершённых заданий по инспекторам
func (s *AnalyticsService) GenerateInspectorPerformancePNG(ctx context.Context, from, to time.Time) ([]byte, error) {
	// Получаем задачи Approved за период с edge Inspector
	tasks, err := s.Client.Task.Query().
		Where(task.StatusEQ(task.StatusApproved), task.CreatedAtGTE(from), task.CreatedAtLTE(to)).
		WithInspector().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	counts := map[string]float64{}
	for _, t := range tasks {
		if t.Edges.Inspector == nil {
			continue
		}
		name := fmt.Sprintf("%s %s", t.Edges.Inspector.FirstName, t.Edges.Inspector.LastName)
		counts[name] += 1
	}

	// Подготовим данные
	labels := make([]string, 0, len(counts))
	vals := make(plotter.Values, 0, len(counts))
	for k, v := range counts {
		labels = append(labels, k)
		vals = append(vals, v)
	}

	p := plot.New()
	p.Title.Text = "Inspector Performance"
	if len(labels) > 0 {
		p.NominalX(labels...)
	}

	if len(vals) > 0 {
		bar, err := plotter.NewBarChart(vals, vg.Points(20))
		if err != nil {
			return nil, err
		}
		p.Add(bar)
	}

	// Render into PNG buffer
	width := vg.Inch * 8
	height := vg.Inch * 4
	img := vgimg.New(width, height)
	dc := draw.New(img)
	p.Draw(dc)

	buf := &bytes.Buffer{}
	pngCanvas := vgimg.PngCanvas{Canvas: img}
	if _, err := pngCanvas.WriteTo(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateStatusDistributionPNG — распределение статусов заданий по районам
func (s *AnalyticsService) GenerateStatusDistributionPNG(ctx context.Context, from, to time.Time) ([]byte, error) {
	// Получаем задания за период с связями Building -> District
	tasks, err := s.Client.Task.Query().
		Where(task.CreatedAtGTE(from), task.CreatedAtLTE(to)).
		WithBuilding(func(bq *ent.BuildingQuery) {
			bq.WithDistrict()
		}).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Группируем: district -> status -> count
	type districtStats struct {
		name   string
		counts map[task.Status]int
	}
	districtMap := make(map[int]*districtStats)

	for _, t := range tasks {
		if t.Edges.Building == nil || t.Edges.Building.Edges.District == nil {
			continue
		}
		d := t.Edges.Building.Edges.District
		if _, ok := districtMap[d.ID]; !ok {
			districtMap[d.ID] = &districtStats{
				name:   d.Name,
				counts: make(map[task.Status]int),
			}
		}
		districtMap[d.ID].counts[t.Status]++
	}

	// Собираем названия районов и статусы
	var districts []*districtStats
	for _, ds := range districtMap {
		districts = append(districts, ds)
	}
	// Сортируем районы по имени для стабильного вывода
	sort.Slice(districts, func(i, j int) bool {
		return districts[i].name < districts[j].name
	})

	// Определяем все возможные статусы
	allStatuses := []task.Status{
		task.StatusNew,
		task.StatusPending,
		task.StatusInProgress,
		task.StatusOnReview,
		task.StatusForRevision,
		task.StatusApproved,
		task.StatusCanceled,
	}

	// Создаём график
	p := plot.New()
	p.Title.Text = "Распределение статусов заданий по районам"
	p.Y.Label.Text = "Количество заданий"

	// Подготовка данных для групповой столбчатой диаграммы
	districtNames := make([]string, len(districts))
	for i, d := range districts {
		districtNames[i] = d.name
	}

	if len(districtNames) > 0 {
		p.NominalX(districtNames...)
	}

	// Цвета для статусов
	statusColors := map[task.Status]color.RGBA{
		task.StatusNew:         {R: 100, G: 149, B: 237, A: 255}, // Cornflower Blue
		task.StatusPending:     {R: 255, G: 215, B: 0, A: 255},   // Gold
		task.StatusInProgress:  {R: 255, G: 165, B: 0, A: 255},   // Orange
		task.StatusOnReview:    {R: 147, G: 112, B: 219, A: 255}, // Medium Purple
		task.StatusForRevision: {R: 255, G: 99, B: 71, A: 255},   // Tomato
		task.StatusApproved:    {R: 50, G: 205, B: 50, A: 255},   // Lime Green
		task.StatusCanceled:    {R: 128, G: 128, B: 128, A: 255}, // Gray
	}

	barWidth := vg.Points(10)
	offset := -float64(len(allStatuses)-1) / 2.0

	for i, status := range allStatuses {
		vals := make(plotter.Values, len(districts))
		hasData := false
		for j, d := range districts {
			vals[j] = float64(d.counts[status])
			if d.counts[status] > 0 {
				hasData = true
			}
		}
		if !hasData && len(districts) == 0 {
			continue
		}

		bar, err := plotter.NewBarChart(vals, barWidth)
		if err != nil {
			continue
		}
		bar.Color = statusColors[status]
		bar.Offset = vg.Points((offset + float64(i)) * 12)
		p.Add(bar)
		p.Legend.Add(string(status), bar)
	}

	p.Legend.Top = true
	p.Legend.Left = false

	// Render into PNG buffer
	width := vg.Inch * 10
	height := vg.Inch * 6
	img := vgimg.New(width, height)
	dc := draw.New(img)
	p.Draw(dc)

	buf := &bytes.Buffer{}
	pngCanvas := vgimg.PngCanvas{Canvas: img}
	if _, err := pngCanvas.WriteTo(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateFailureFrequencyPNG — частота "Аварийных" и "Неудовлетворительных" статусов по элементам
func (s *AnalyticsService) GenerateFailureFrequencyPNG(ctx context.Context, from, to time.Time) ([]byte, error) {
	// Получаем результаты осмотра за период
	results, err := s.Client.InspectionResult.Query().
		Where(
			inspectionresult.Or(
				inspectionresult.ConditionStatusEQ(inspectionresult.ConditionStatusАварийное),
				inspectionresult.ConditionStatusEQ(inspectionresult.ConditionStatusНеудовлетворительное),
			),
		).
		WithTask(func(tq *ent.TaskQuery) {
			tq.Where(task.CreatedAtGTE(from), task.CreatedAtLTE(to))
		}).
		WithChecklistElement(func(ceq *ent.ChecklistElementQuery) {
			ceq.WithElementCatalog()
		}).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Фильтруем только те, у которых task в нужном периоде
	type elementStats struct {
		name                  string
		unsatisfactoryCount   int
		emergencyCount        int
	}
	elementMap := make(map[int]*elementStats)

	for _, r := range results {
		// Проверяем что task загружен (фильтр по дате)
		if r.Edges.Task == nil {
			continue
		}
		if r.Edges.ChecklistElement == nil || r.Edges.ChecklistElement.Edges.ElementCatalog == nil {
			continue
		}

		elemCatalog := r.Edges.ChecklistElement.Edges.ElementCatalog
		if _, ok := elementMap[elemCatalog.ID]; !ok {
			elementMap[elemCatalog.ID] = &elementStats{name: elemCatalog.Name}
		}

		switch r.ConditionStatus {
		case inspectionresult.ConditionStatusНеудовлетворительное:
			elementMap[elemCatalog.ID].unsatisfactoryCount++
		case inspectionresult.ConditionStatusАварийное:
			elementMap[elemCatalog.ID].emergencyCount++
		}
	}

	// Собираем и сортируем по общему количеству проблем
	var elements []*elementStats
	for _, es := range elementMap {
		elements = append(elements, es)
	}
	sort.Slice(elements, func(i, j int) bool {
		totalI := elements[i].unsatisfactoryCount + elements[i].emergencyCount
		totalJ := elements[j].unsatisfactoryCount + elements[j].emergencyCount
		return totalI > totalJ // Сортируем по убыванию
	})

	// Ограничиваем топ-15 элементов для читаемости
	if len(elements) > 15 {
		elements = elements[:15]
	}

	// Создаём график
	p := plot.New()
	p.Title.Text = "Частота проблемных состояний по элементам"
	p.Y.Label.Text = "Количество"

	elementNames := make([]string, len(elements))
	for i, e := range elements {
		elementNames[i] = e.name
	}

	if len(elementNames) > 0 {
		p.NominalX(elementNames...)
	}

	// Данные для "Неудовлетворительное"
	unsatisfactoryVals := make(plotter.Values, len(elements))
	emergencyVals := make(plotter.Values, len(elements))
	for i, e := range elements {
		unsatisfactoryVals[i] = float64(e.unsatisfactoryCount)
		emergencyVals[i] = float64(e.emergencyCount)
	}

	barWidth := vg.Points(15)

	// Столбцы "Неудовлетворительное" — оранжевые
	if len(unsatisfactoryVals) > 0 {
		barUnsatisfactory, err := plotter.NewBarChart(unsatisfactoryVals, barWidth)
		if err == nil {
			barUnsatisfactory.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
			barUnsatisfactory.Offset = vg.Points(-8)
			p.Add(barUnsatisfactory)
			p.Legend.Add("Неудовлетворительное", barUnsatisfactory)
		}
	}

	// Столбцы "Аварийное" — красные
	if len(emergencyVals) > 0 {
		barEmergency, err := plotter.NewBarChart(emergencyVals, barWidth)
		if err == nil {
			barEmergency.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255} // Crimson
			barEmergency.Offset = vg.Points(8)
			p.Add(barEmergency)
			p.Legend.Add("Аварийное", barEmergency)
		}
	}

	p.Legend.Top = true

	// Render into PNG buffer
	width := vg.Inch * 10
	height := vg.Inch * 5
	img := vgimg.New(width, height)
	dc := draw.New(img)
	p.Draw(dc)

	buf := &bytes.Buffer{}
	pngCanvas := vgimg.PngCanvas{Canvas: img}
	if _, err := pngCanvas.WriteTo(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateReportPDF — сборка PDF с графиками
func (s *AnalyticsService) GenerateReportPDF(ctx context.Context, from, to time.Time, charts []string) ([]byte, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Подключаем шрифты, если указаны
	pdf.AddUTF8Font("Times", "", "storage/fonts/timesnewromanpsmt.ttf")  // Путь к обычному шрифту
    if err := pdf.Error(); err != nil {
        return nil, "", fmt.Errorf("failed to load regular font: %w", err)
    }
    pdf.AddUTF8Font("Times", "B", "storage/fonts/ofont.ru_Times New Roman.ttf")  // Путь к жирному шрифту
    if err := pdf.Error(); err != nil {
        return nil, "", fmt.Errorf("failed to load bold font: %w", err)
    }

	// Титульная страница
	pdf.AddPage()
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(0, 10, "Аналитический отчёт", "", 0, "C", false, 0, "")
	pdf.Ln(12)
	pdf.SetFont("Times", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Период: %s — %s", from.Format("02.01.2006"), to.Format("02.01.2006")), "", 1, "L", false, 0, "")

	// Маппинг названий графиков для PDF
	chartTitles := map[string]string{
		"inspector_performance": "Производительность инспекторов",
		"status_distribution":   "Распределение статусов заданий по районам",
		"failure_frequency":     "Частота проблемных состояний элементов",
	}

	for _, ch := range charts {
		var img []byte
		var err error

		switch ch {
		case "inspector_performance":
			img, err = s.GenerateInspectorPerformancePNG(ctx, from, to)
		case "status_distribution":
			img, err = s.GenerateStatusDistributionPNG(ctx, from, to)
		case "failure_frequency":
			img, err = s.GenerateFailureFrequencyPNG(ctx, from, to)
		default:
			// Пропускаем неподдерживаемые
			continue
		}

		if err != nil {
			return nil, "", fmt.Errorf("failed to generate chart %s: %w", ch, err)
		}

		// Встраиваем график в PDF
		name := fmt.Sprintf("chart_%s", ch)
		pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(img))
		pdf.AddPage()
		pdf.SetFont("Times", "B", 14)
		title := chartTitles[ch]
		if title == "" {
			title = ch
		}
		pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")
		pdf.ImageOptions(name, 10, 30, 190, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	buf := &bytes.Buffer{}
	if err := pdf.Output(buf); err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
	}
	filename := fmt.Sprintf("analytics_%s_%s.pdf", from.Format("20060102"), to.Format("20060102"))
	return buf.Bytes(), filename, nil
}
