//service

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"jkh/ent"
	"jkh/ent/checklistelement"
	"jkh/ent/inspectionresult"
	"jkh/ent/task"
	"jkh/pkg/models"
)

// ============================================================================
// ОШИБКИ БИЗНЕС-ЛОГИКИ
// ============================================================================

var (
	ErrResultNotFound          = errors.New("inspection result not found")
	ErrResultAlreadyExists     = errors.New("result for this element already exists")
	ErrTaskNotInProgress       = errors.New("task is not in progress (cannot add results)")
	ErrChecklistElementInvalid = errors.New("checklist element does not belong to task's checklist")
)

// ============================================================================
// СЕРВИС
// ============================================================================

type InspectionResultService struct {
	Client *ent.Client
}

func NewInspectionResultService(client *ent.Client) *InspectionResultService {
	return &InspectionResultService{Client: client}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// toInspectionResultResponse — преобразование Ent → DTO.
func (s *InspectionResultService) toInspectionResultResponse(ir *ent.InspectionResult) *models.InspectionResultResponse {
	resp := &models.InspectionResultResponse{
		TaskID:             ir.TaskID,
		ChecklistElementID: ir.ChecklistElementID,
		ConditionStatus:    string(ir.ConditionStatus),
		Comment:            ir.Comment,
		CreatedAt:          ir.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          ir.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Если загружен ChecklistElement → ElementCatalog, добавляем информацию
	if ir.Edges.ChecklistElement != nil {
		ce := ir.Edges.ChecklistElement
		resp.OrderIndex = ce.OrderIndex

		if ce.Edges.ElementCatalog != nil {
			resp.ElementName = ce.Edges.ElementCatalog.Name
			resp.ElementCategory = ce.Edges.ElementCatalog.Category
		}
	}

	return resp
}

// validateTaskAndElement — проверка, что задание в статусе InProgress и элемент принадлежит чек-листу.
func (s *InspectionResultService) validateTaskAndElement(ctx context.Context, taskID, checklistElementID int) error {
	// 1. Получаем задание с чек-листом
	t, err := s.Client.Task.Query().
		Where(task.IDEQ(taskID)).
		WithChecklist().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("database error: %w", err)
	}

	// 2. Проверяем, что задание в статусе InProgress
	if t.Status != task.StatusInProgress {
		return ErrTaskNotInProgress
	}

	// 3. Проверяем, что ChecklistElement принадлежит чек-листу задания
	exists, err := s.Client.ChecklistElement.Query().
		Where(
			checklistelement.IDEQ(checklistElementID),
			checklistelement.ChecklistIDEQ(t.ChecklistID),
		).
		Exist(ctx)

	if err != nil || !exists {
		return ErrChecklistElementInvalid
	}

	return nil
}

// ============================================================================
// CRUD-ОПЕРАЦИИ
// ============================================================================

// CreateOrUpdateResult — создание или обновление результата проверки элемента.
// Если результат уже существует (task_id + checklist_element_id), обновляем его.
func (s *InspectionResultService) CreateOrUpdateResult(ctx context.Context, taskID int, req models.CreateInspectionResultRequest) (*models.InspectionResultResponse, error) {
	// 1. Валидация
	if err := s.validateTaskAndElement(ctx, taskID, req.ChecklistElementID); err != nil {
		return nil, err
	}

	// 2. Проверяем, существует ли уже результат
	existing, err := s.Client.InspectionResult.Query().
		Where(
			inspectionresult.TaskIDEQ(taskID),
			inspectionresult.ChecklistElementIDEQ(req.ChecklistElementID),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	var result *ent.InspectionResult

	if existing != nil {
		// Обновление существующего результата
		update := s.Client.InspectionResult.UpdateOne(existing).
			SetConditionStatus(inspectionresult.ConditionStatus(req.ConditionStatus))

		if req.Comment != nil {
			update.SetComment(*req.Comment)
		} else {
			update.ClearComment()
		}

		_, err = update.Save(ctx)
		if err != nil {
			log.Printf("DB error updating inspection result: %v", err)
			return nil, fmt.Errorf("database error")
		}
	} else {
		// Создание нового результата
		create := s.Client.InspectionResult.Create().
			SetTaskID(taskID).
			SetChecklistElementID(req.ChecklistElementID).
			SetConditionStatus(inspectionresult.ConditionStatus(req.ConditionStatus))

		if req.Comment != nil {
			create.SetComment(*req.Comment)
		}

		_, err = create.Save(ctx)
		if err != nil {
			log.Printf("DB error creating inspection result: %v", err)
			return nil, fmt.Errorf("database error")
		}
	}

	// 3. Догружаем связи для ответа
	result, err = s.Client.InspectionResult.Query().
		Where(
			inspectionresult.TaskIDEQ(taskID),
			inspectionresult.ChecklistElementIDEQ(req.ChecklistElementID),
		).
		WithChecklistElement(func(q *ent.ChecklistElementQuery) {
			q.WithElementCatalog()
		}).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch result: %w", err)
	}

	return s.toInspectionResultResponse(result), nil
}

// GetTaskResults — получение всех результатов для задания (сводка).
func (s *InspectionResultService) GetTaskResults(ctx context.Context, taskID int) (*models.TaskResultsSummary, error) {
	// 1. Получаем задание с чек-листом и элементами
	t, err := s.Client.Task.Query().
		Where(task.IDEQ(taskID)).
		WithChecklist(func(q *ent.ChecklistQuery) {
			q.WithElements(func(ceq *ent.ChecklistElementQuery) {
				ceq.WithElementCatalog()
			})
		}).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Получаем все результаты для задания
	results, err := s.Client.InspectionResult.Query().
		Where(inspectionresult.TaskIDEQ(taskID)).
		WithChecklistElement(func(q *ent.ChecklistElementQuery) {
			q.WithElementCatalog()
		}).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 3. Формируем ответ
	summary := &models.TaskResultsSummary{
		TaskID:            t.ID,
		TaskTitle:         t.Title,
		TotalElements:     len(t.Edges.Checklist.Edges.Elements),
		CompletedElements: len(results),
		Results:           []models.InspectionResultResponse{},
	}

	for _, r := range results {
		summary.Results = append(summary.Results, *s.toInspectionResultResponse(r))
	}

	return summary, nil
}

// DeleteResult — удаление результата проверки элемента.
func (s *InspectionResultService) DeleteResult(ctx context.Context, taskID, checklistElementID int) error {
	deleted, err := s.Client.InspectionResult.Delete().
		Where(
			inspectionresult.TaskIDEQ(taskID),
			inspectionresult.ChecklistElementIDEQ(checklistElementID),
		).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if deleted == 0 {
		return ErrResultNotFound
	}

	return nil
}
