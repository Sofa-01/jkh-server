// service/inspectionact.go

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"jkh/ent"
	"jkh/ent/inspectionact"
	"jkh/ent/inspectionresult"

	"github.com/jung-kurt/gofpdf"
)

// ============================================================================
// ОШИБКИ
// ============================================================================

var (
	ErrActNotFound = errors.New("inspection act not found")
)

// ============================================================================
// СЕРВИС
// ============================================================================

type InspectionActService struct {
	Client      *ent.Client
	StoragePath string // Путь для сохранения PDF (например, "storage/acts")
}

func NewInspectionActService(client *ent.Client, storagePath string) *InspectionActService {
	// Создаём директорию, если её нет
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Printf("failed to create storage directory %s: %v", storagePath, err)
	}
	return &InspectionActService{
		Client:      client,
		StoragePath: storagePath,
	}
}

// ============================================================================
// СОЗДАНИЕ / ОБНОВЛЕНИЕ ЗАПИСИ АКТА
// ============================================================================

// CreateOrUpdateAct — создаёт или обновляет запись акта для задания.
// Вызывается, когда инспектор отправляет задание на проверку (InProgress → OnReview).
func (s *InspectionActService) CreateOrUpdateAct(ctx context.Context, taskID int, conclusion string) (*ent.InspectionAct, error) {
	// Проверяем, есть ли уже акт для этого задания
	act, err := s.Client.InspectionAct.Query().
		Where(inspectionact.TaskIDEQ(taskID)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if act != nil {
        //  Просто обновляем conclusion, PDF не трогаем
        act, err = s.Client.InspectionAct.UpdateOne(act).
            SetConclusion(conclusion).
            Save(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to update inspection act: %w", err)
        }
        return act, nil
    }

	// Создаём новый акт
	act, err = s.Client.InspectionAct.Create().
		SetTaskID(taskID).
		SetStatus("создан").
		SetConclusion(conclusion).
		Save(ctx)

	if err != nil {
		log.Printf("DB error creating inspection act: %v", err)
		return nil, fmt.Errorf("database error")
	}

	return act, nil
}

// ApproveAct — Перегенерировать PDF с датой утверждения
func (s *InspectionActService) ApproveAct(ctx context.Context, taskID int) error {
    //1. Загружаем акт со всеми связями
	act, err := s.Client.InspectionAct.Query().
        Where(inspectionact.TaskIDEQ(taskID)).
        WithTask(func(tq *ent.TaskQuery) {
            tq.WithBuilding(func(bq *ent.BuildingQuery) {
                bq.WithDistrict().WithJkhUnit()
            }).
            WithChecklist(func(cq *ent.ChecklistQuery) {
                cq.WithElements(func(ceq *ent.ChecklistElementQuery) {
                    ceq.WithElementCatalog()
                })
            }).
            WithInspector()
        }).
        Only(ctx)

    if err != nil {
        if ent.IsNotFound(err) {
            return ErrActNotFound
        }
        return fmt.Errorf("database error: %w", err)
    }

    now := time.Now()
    
    // 2. Обновляем статус, дату И заключение в БД
    _, err = s.Client.InspectionAct.UpdateOne(act).
        SetApprovedAt(now).
        SetStatus("утверждён").
        SetConclusion("Акт осмотра утверждён координатором.").  // ← ✅ ДОБАВЛЕНО!
        Save(ctx)
        
    if err != nil {
        return fmt.Errorf("failed to approve inspection act: %w", err)
    }

	// 3. Обновляем act вручную (для generatePDF)
	act.ApprovedAt = now
    act.Status = "утверждён"
    act.Conclusion = "Акт осмотра утверждён координатором."

    // 4. Удаляем старый PDF (черновик)
    if act.DocumentPath != "" {
        if err := os.Remove(act.DocumentPath); err != nil {
            log.Printf("failed to delete draft PDF: %v", err)
        }
    }

    // 5. Получаем результаты осмотра
    results, err := s.Client.InspectionResult.Query().
        Where(inspectionresult.TaskIDEQ(taskID)).
        WithChecklistElement(func(ceq *ent.ChecklistElementQuery) {
            ceq.WithElementCatalog()
        }).
        All(ctx)

    if err != nil {
        log.Printf("failed to fetch results for approved PDF: %v", err)
        return nil // Не критично, основная задача выполнена
    }

    // 6. Генерируем ФИНАЛЬНЫЙ утверждённый PDF
    pdfData, filename, err := s.generatePDF(act, results)
    if err != nil {
        log.Printf("failed to generate approved PDF: %v", err)
        return nil // Не критично
    }

    // 7. Сохраняем финальный PDF
    fullPath := filepath.Join(s.StoragePath, filename)
    if err := os.WriteFile(fullPath, pdfData, 0644); err != nil {
        log.Printf("failed to save approved PDF: %v", err)
        return nil
    }

    // 6. Обновляем document_path
    _, err = s.Client.InspectionAct.UpdateOne(act).
        SetDocumentPath(fullPath).
        Save(ctx)
    if err != nil {
        log.Printf("failed to update document_path: %v", err)
    }

    log.Printf("Approved PDF generated for task %d", taskID)
    return nil
}


// ============================================================================
// ГЕНЕРАЦИЯ / ВОЗВРАТ PDF
// ============================================================================

// GeneratePDFForAct — генерирует PDF (если нужно) и возвращает []byte + имя файла.
// Если document_path уже заполнен и файл существует — просто читает его.
func (s *InspectionActService) GeneratePDFForAct(ctx context.Context, taskID int) ([]byte, string, error) {
	// 1. Получаем акт вместе с задачей
	act, err := s.Client.InspectionAct.Query().
		Where(inspectionact.TaskIDEQ(taskID)).
		WithTask(func(tq *ent.TaskQuery) {
			tq.
				WithBuilding(func(bq *ent.BuildingQuery) {
					bq.WithDistrict().WithJkhUnit()
				}).
				WithChecklist(func(cq *ent.ChecklistQuery) {
					cq.WithElements(func(ceq *ent.ChecklistElementQuery) {
						ceq.WithElementCatalog()
					})
				}).
				WithInspector()
		}).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", ErrActNotFound
		}
		return nil, "", fmt.Errorf("database error: %w", err)
	}

	// 2. Если PDF уже есть на диске — читаем и возвращаем
	if act.DocumentPath != "" {
		data, err := os.ReadFile(act.DocumentPath)
		if err == nil {
			filename := filepath.Base(act.DocumentPath)
			return data, filename, nil
		}
		log.Printf("failed to read existing PDF, will regenerate: %v", err)
	}

	// 3. Получаем результаты осмотра
	results, err := s.Client.InspectionResult.Query().
		Where(inspectionresult.TaskIDEQ(taskID)).
		WithChecklistElement(func(ceq *ent.ChecklistElementQuery) {
			ceq.WithElementCatalog()
		}).
		All(ctx)

	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch inspection results: %w", err)
	}

	// 4. Генерируем PDF в памяти
	pdfData, filename, err := s.generatePDF(act, results)
	if err != nil {
		return nil, "", err
	}

	// 5. Сохраняем PDF на диск
	fullPath := filepath.Join(s.StoragePath, filename)
	if err := os.WriteFile(fullPath, pdfData, 0644); err != nil {
		log.Printf("failed to save PDF to %s: %v", fullPath, err)
		return nil, "", fmt.Errorf("failed to save PDF")
	}

	// 6. Обновляем document_path в акте
	_, err = s.Client.InspectionAct.UpdateOne(act).
		SetDocumentPath(fullPath).
		Save(ctx)
	if err != nil {
		log.Printf("failed to update inspection act with document_path: %v", err)
	}

	return pdfData, filename, nil
}

// ============================================================================
// ВНУТРЕННЯЯ ГЕНЕРАЦИЯ PDF
// ============================================================================

func (s *InspectionActService) generatePDF(act *ent.InspectionAct, results []*ent.InspectionResult) ([]byte, string, error) {
    t := act.Edges.Task
    if t == nil {
        return nil, "", fmt.Errorf("task edge not loaded for inspection act")
    }

    pdf := gofpdf.New("P", "mm", "A4", "")
    // Подключаем шрифт с кириллицей (предполагаем, что файл .ttf лежит по этому пути)
    // Загружаем шрифты Times New Roman и проверяем ошибки
    pdf.AddUTF8Font("Times", "", "storage/fonts/timesnewromanpsmt.ttf")  // Путь к обычному шрифту
    if err := pdf.Error(); err != nil {
        return nil, "", fmt.Errorf("failed to load regular font: %w", err)
    }
    pdf.AddUTF8Font("Times", "B", "storage/fonts/ofont.ru_Times New Roman.ttf")  // Путь к жирному шрифту
    if err := pdf.Error(); err != nil {
        return nil, "", fmt.Errorf("failed to load bold font: %w", err)
    }
    pdf.AddPage()

    // Заголовок
    pdf.SetFont("Times", "B", 16)  // Теперь используем "Times" вместо "DejaVu"
    pdf.CellFormat(0, 10, "АКТ ОСМОТРА ЖИЛОГО ПОМЕЩЕНИЯ", "", 0, "C", false, 0, "")
    pdf.Ln(13)

    // Основная информация
    pdf.SetFont("Times", "B", 12)
    pdf.CellFormat(0, 8, "ОСНОВНАЯ ИНФОРМАЦИЯ", "", 0, "L", false, 0, "")
    pdf.Ln(8)
    pdf.SetFont("Times", "", 11)

    pdf.CellFormat(55, 6, "Номер акта:", "", 0, "L", false, 0, "")
    pdf.CellFormat(0, 6, fmt.Sprintf("%d", act.ID), "", 0, "L", false, 0, "")
    pdf.Ln(6)

    pdf.CellFormat(55, 6, "Дата создания акта:", "", 0, "L", false, 0, "")
    pdf.CellFormat(0, 6, act.CreatedAt.Format("02.01.2006 15:04"), "", 0, "L", false, 0, "")
    pdf.Ln(6)

    pdf.CellFormat(55, 6, "Статус акта:", "", 0, "L", false, 0, "")
    pdf.CellFormat(0, 6, act.Status, "", 0, "L", false, 0, "")
    pdf.Ln(6)

    if !act.ApprovedAt.IsZero() {
        pdf.CellFormat(55, 6, "Дата утверждения:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, act.ApprovedAt.Format("02.01.2006 15:04"), "", 0, "L", false, 0, "")
        pdf.Ln(6)
    }

    pdf.CellFormat(55, 6, "Дата осмотра:", "", 0, "L", false, 0, "")
    pdf.CellFormat(0, 6, t.ScheduledDate.Format("02.01.2006"), "", 0, "L", false, 0, "")
    pdf.Ln(6)

    if t.Edges.Inspector != nil {
        ins := t.Edges.Inspector
        pdf.CellFormat(55, 6, "Инспектор:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, fmt.Sprintf("%s %s", ins.FirstName, ins.LastName), "", 0, "L", false, 0, "")
        pdf.Ln(6)

        pdf.CellFormat(55, 6, "Email инспектора:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, ins.Email, "", 0, "L", false, 0, "")
        pdf.Ln(6)
    }

    pdf.Ln(3)

    // Информация о здании
    pdf.SetFont("Times", "B", 12)
    pdf.CellFormat(0, 8, "ИНФОРМАЦИЯ О ЗДАНИИ", "", 0, "L", false, 0, "")
    pdf.Ln(8)
    pdf.SetFont("Times", "", 11)

    if t.Edges.Building != nil {
        b := t.Edges.Building

        pdf.CellFormat(55, 6, "Адрес:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, b.Address, "", 0, "L", false, 0, "")
        pdf.Ln(6)

        pdf.CellFormat(55, 6, "Год постройки:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, fmt.Sprintf("%d", b.ConstructionYear), "", 0, "L", false, 0, "")
        pdf.Ln(6)

        if b.Edges.District != nil {
            pdf.CellFormat(55, 6, "Район:", "", 0, "L", false, 0, "")
            pdf.CellFormat(0, 6, b.Edges.District.Name, "", 0, "L", false, 0, "")
            pdf.Ln(6)
        }
        if b.Edges.JkhUnit != nil {
            pdf.CellFormat(55, 6, "ЖКХ:", "", 0, "L", false, 0, "")
            pdf.CellFormat(0, 6, b.Edges.JkhUnit.Name, "", 0, "L", false, 0, "")
            pdf.Ln(6)
        }
    }

    pdf.Ln(3)

    // Чек-лист
    pdf.SetFont("Times", "B", 12)
    pdf.CellFormat(0, 8, "ЧЕК-ЛИСТ ОСМОТРА", "", 0, "L", false, 0, "")
    pdf.Ln(8)
    pdf.SetFont("Times", "", 11)

    if t.Edges.Checklist != nil {
        cl := t.Edges.Checklist
        pdf.CellFormat(55, 6, "Название:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, cl.Title, "", 0, "L", false, 0, "")
        pdf.Ln(6)

        pdf.CellFormat(55, 6, "Тип осмотра:", "", 0, "L", false, 0, "")
        pdf.CellFormat(0, 6, string(cl.InspectionType), "", 0, "L", false, 0, "")
        pdf.Ln(6)
    }

    pdf.Ln(3)

    // Таблица результатов
    pdf.SetFont("Times", "B", 12)
    pdf.CellFormat(0, 8, "РЕЗУЛЬТАТЫ ОСМОТРА", "", 0, "L", false, 0, "")
    pdf.Ln(8)
    pdf.SetFont("Times", "", 9)

    // Заголовки таблицы
    pdf.SetFillColor(220, 220, 220)
    pdf.CellFormat(10, 7, "№", "1", 0, "C", true, 0, "")
    pdf.CellFormat(45, 7, "Элемент", "1", 0, "L", true, 0, "")
    pdf.CellFormat(40, 7, "Статус", "1", 0, "L", true, 0, "")
    pdf.CellFormat(95, 7, "Примечание", "1", 1, "L", true, 0, "")
    pdf.SetFillColor(255, 255, 255)

    for i, r := range results {
        elemName := ""
        if r.Edges.ChecklistElement != nil && r.Edges.ChecklistElement.Edges.ElementCatalog != nil {
            elemName = r.Edges.ChecklistElement.Edges.ElementCatalog.Name
        }

        pdf.CellFormat(10, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
        pdf.CellFormat(45, 6, elemName, "1", 0, "L", false, 0, "")
        pdf.CellFormat(40, 6, string(r.ConditionStatus), "1", 0, "L", false, 0, "")
        pdf.CellFormat(95, 6, r.Comment, "1", 1, "L", false, 0, "")
    }

    pdf.Ln(4)

    // Заключение
    pdf.SetFont("Times", "B", 12)
    pdf.CellFormat(0, 8, "ЗАКЛЮЧЕНИЕ", "", 0, "L", false, 0, "")
    pdf.Ln(8)
    pdf.SetFont("Times", "", 10)

    conclusion := act.Conclusion
    if conclusion == "" {
        conclusion = "Осмотр выполнен. Результаты представлены в таблице выше."
    }
    pdf.MultiCell(0, 5, conclusion, "", "L", false)
    pdf.Ln(8)

    // Подпись
    pdf.SetFont("Times", "", 10)
    pdf.CellFormat(90, 6, "Подпись инспектора: ____________________", "", 0, "L", false, 0, "")
    pdf.CellFormat(0, 6, "Дата: "+time.Now().Format("02.01.2006"), "", 1, "L", false, 0, "")

    buf := new(bytes.Buffer)
    if err := pdf.Output(buf); err != nil {
        return nil, "", fmt.Errorf("failed to generate PDF: %w", err)
    }

    filename := fmt.Sprintf("act_%d_%s.pdf", act.TaskID, time.Now().Format("20060102_150405"))
    return buf.Bytes(), filename, nil
}



