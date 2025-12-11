package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"jkh/ent"
	"jkh/ent/building"
	"jkh/ent/checklist"
	"jkh/ent/inspectorunit"
	"jkh/ent/task"
	"jkh/ent/user"
	"jkh/pkg/models"
)

// ============================================================================
// ОШИБКИ БИЗНЕС-ЛОГИКИ
// ============================================================================

var (
	ErrTaskNotFound            = errors.New("task not found")
	ErrInvalidForeignKey       = errors.New("invalid building, checklist, or inspector ID")
	ErrInspectorNotAssigned    = errors.New("inspector not assigned to building's JKH unit")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrUnauthorizedAction      = errors.New("unauthorized to perform this action")
)

// ============================================================================
// FSM — КОНЕЧНЫЙ АВТОМАТ СОСТОЯНИЙ
// ============================================================================

// allowedTransitions определяет разрешенные переходы между статусами.
var allowedTransitions = map[task.Status][]task.Status{
	task.StatusNew:         {task.StatusPending, task.StatusCanceled},
	task.StatusPending:     {task.StatusInProgress, task.StatusCanceled},
	task.StatusInProgress:  {task.StatusOnReview, task.StatusCanceled},
	task.StatusOnReview:    {task.StatusApproved, task.StatusForRevision},
	task.StatusForRevision: {task.StatusOnReview, task.StatusCanceled},
	task.StatusApproved:    {}, // Финальное состояние
	task.StatusCanceled:    {}, // Финальное состояние
}

// isTransitionAllowed проверяет, разрешен ли переход из currentStatus в newStatus.
func isTransitionAllowed(currentStatus, newStatus task.Status) bool {
	allowed, exists := allowedTransitions[currentStatus]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// ============================================================================
// СЕРВИС
// ============================================================================

type TaskService struct {
	Client *ent.Client
}

func NewTaskService(client *ent.Client) *TaskService {
	return &TaskService{Client: client}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// toTaskResponse — преобразование Ent → базовый DTO.
func (s *TaskService) toTaskResponse(t *ent.Task) *models.TaskResponse {
	resp := &models.TaskResponse{
		ID:            t.ID,
		Title:         t.Title,
		Status:        string(t.Status),
		Priority:      t.Priority,
		ScheduledDate: t.ScheduledDate.Format(time.RFC3339),
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
	}

	// Добавляем информацию о связанных сущностях
	if t.Edges.Building != nil {
		resp.BuildingAddress = t.Edges.Building.Address
	}
	if t.Edges.Checklist != nil {
		resp.ChecklistTitle = t.Edges.Checklist.Title
	}
	if t.Edges.Inspector != nil {
		resp.InspectorName = fmt.Sprintf("%s %s",
			t.Edges.Inspector.FirstName,
			t.Edges.Inspector.LastName)
	}

	return resp
}

// toTaskDetailResponse — преобразование Ent → детальный DTO.
func (s *TaskService) toTaskDetailResponse(t *ent.Task) *models.TaskDetailResponse {
	resp := &models.TaskDetailResponse{
		ID:            t.ID,
		Title:         t.Title,
		Status:        string(t.Status),
		Priority:      t.Priority,
		Description:   t.Description,
		ScheduledDate: t.ScheduledDate.Format(time.RFC3339),
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
	}

	// Заполняем детальную информацию о связанных сущностях
	if t.Edges.Building != nil {
		resp.Building = models.BuildingInfo{
			ID:      t.Edges.Building.ID,
			Address: t.Edges.Building.Address,
		}
	}
	if t.Edges.Checklist != nil {
		resp.Checklist = models.ChecklistInfo{
			ID:             t.Edges.Checklist.ID,
			Title:          t.Edges.Checklist.Title,
			InspectionType: string(t.Edges.Checklist.InspectionType),
		}
	}
	if t.Edges.Inspector != nil {
		resp.Inspector = models.InspectorInfo{
			ID:        t.Edges.Inspector.ID,
			FirstName: t.Edges.Inspector.FirstName,
			LastName:  t.Edges.Inspector.LastName,
			Email:     t.Edges.Inspector.Email,
		}
	}

	return resp
}

// validateForeignKeys — проверка существования связанных сущностей.
func (s *TaskService) validateForeignKeys(ctx context.Context, buildingID, checklistID, inspectorID int) error {
	// Проверка Building
	bExists, err := s.Client.Building.Query().Where(building.IDEQ(buildingID)).Exist(ctx)
	if err != nil || !bExists {
		return ErrInvalidForeignKey
	}

	// Проверка Checklist
	cExists, err := s.Client.Checklist.Query().Where(checklist.IDEQ(checklistID)).Exist(ctx)
	if err != nil || !cExists {
		return ErrInvalidForeignKey
	}

	// Проверка Inspector (User)
	iExists, err := s.Client.User.Query().Where(user.IDEQ(inspectorID)).Exist(ctx)
	if err != nil || !iExists {
		return ErrInvalidForeignKey
	}

	return nil
}

// ============================================================================
// CRUD-ОПЕРАЦИИ
// ============================================================================

// CreateTask — создание нового задания (доступно для Coordinator и Specialist).
func (s *TaskService) CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.TaskDetailResponse, error) {
	// 1. Валидация FK
	if err := s.validateForeignKeys(ctx, req.BuildingID, req.ChecklistID, req.InspectorID); err != nil {
		return nil, err
	}

	// 1.1. Проверка, что инспектор закреплён за JKH unit здания
	b, err := s.Client.Building.Query().Where(building.IDEQ(req.BuildingID)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	// Если у здания нет привязанного JKH unit — запрещаем создание задания
	if b.JkhUnitID == 0 {
		return nil, fmt.Errorf("building has no JKH unit assigned")
	}

	assigned, err := s.Client.InspectorUnit.Query().Where(
		inspectorunit.UserIDEQ(req.InspectorID),
		inspectorunit.JkhUnitIDEQ(b.JkhUnitID),
	).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if !assigned {
		return nil, ErrInspectorNotAssigned
	}

	// 2. Парсинг даты
	scheduledDate, err := time.Parse(time.RFC3339, req.ScheduledDate)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_date format (use ISO 8601)")
	}

	// 3. Установка приоритета по умолчанию
	priority := req.Priority
	if priority == "" {
		priority = "обычный"
	}

	// 4. Создание задания
	create := s.Client.Task.Create().
		SetBuildingID(req.BuildingID).
		SetChecklistID(req.ChecklistID).
		SetInspectorID(req.InspectorID).
		SetTitle(req.Title).
		SetPriority(priority).
		SetScheduledDate(scheduledDate).
		SetStatus(task.StatusNew) // Начальный статус

	if req.Description != nil {
		create.SetDescription(*req.Description)
	}

	t, err := create.Save(ctx)
	if err != nil {
		log.Printf("DB error creating task: %v", err)
		return nil, fmt.Errorf("database error")
	}

	// 5. Догружаем связи для ответа
	t, err = s.Client.Task.Query().
		Where(task.IDEQ(t.ID)).
		WithBuilding().
		WithChecklist().
		WithInspector().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created task: %w", err)
	}

	return s.toTaskDetailResponse(t), nil
}

// ListTasks — получение списка заданий.
// Параметры:
//   - inspectorID: если указан, возвращаются только задания этого инспектора
//   - status: фильтр по статусу (опционально)
func (s *TaskService) ListTasks(ctx context.Context, inspectorID *int, status *string) ([]*models.TaskResponse, error) {
	query := s.Client.Task.Query().
		WithBuilding().
		WithChecklist().
		WithInspector()

	// Фильтр по инспектору (для Inspector-роли)
	if inspectorID != nil {
		query = query.Where(task.InspectorIDEQ(*inspectorID))
	}

	// Фильтр по статусу
	if status != nil {
		query = query.Where(task.StatusEQ(task.Status(*status)))
	}

	tasks, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error")
	}

	resp := make([]*models.TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = s.toTaskResponse(t)
	}

	return resp, nil
}

// RetrieveTask — получение детальной информации о задании.
func (s *TaskService) RetrieveTask(ctx context.Context, id int) (*models.TaskDetailResponse, error) {
	t, err := s.Client.Task.Query().
		Where(task.IDEQ(id)).
		WithBuilding().
		WithChecklist().
		WithInspector().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return s.toTaskDetailResponse(t), nil
}

// UpdateTaskStatus — изменение статуса задания (с проверкой FSM).
func (s *TaskService) UpdateTaskStatus(ctx context.Context, id int, newStatus task.Status) error {
	// 1. Получение текущего задания
	t, err := s.Client.Task.Query().Where(task.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("database error: %w", err)
	}

	// 2. Проверка разрешенности перехода
	if !isTransitionAllowed(t.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	// 3. Обновление статуса
	err = s.Client.Task.UpdateOneID(id).
		SetStatus(newStatus).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// ============================================================================
	// ИНТЕГРАЦИЯ С INSPECTION ACT
	// ============================================================================

	// 4. Если переход в OnReview — создаём акт осмотра
	if newStatus == task.StatusOnReview {
		actService := NewInspectionActService(s.Client, "storage/acts")
		conclusion := "Осмотр выполнен. Ожидает проверки координатором."
		_, err := actService.CreateOrUpdateAct(ctx, id, conclusion)
		if err != nil {
			log.Printf("Failed to create inspection act for task %d: %v", id, err)
			// Не прерываем выполнение — акт можно создать позже вручную
		} else {
			log.Printf("Inspection act created for task %d", id)
		}
	}

	// 5. Если переход в Approved — утверждаем акт
	if newStatus == task.StatusApproved {
		actService := NewInspectionActService(s.Client, "storage/acts")
		err := actService.ApproveAct(ctx, id)
		if err != nil {
			log.Printf("Failed to approve inspection act for task %d: %v", id, err)
			// Не критично, продолжаем
		} else {
			log.Printf("Inspection act approved for task %d", id)
		}
	}

	return nil
}

// AssignInspector — переназначение инспектора (только для Coordinator/Specialist).
func (s *TaskService) AssignInspector(ctx context.Context, taskID, inspectorID int) error {
	// Проверка существования инспектора
	exists, err := s.Client.User.Query().Where(user.IDEQ(inspectorID)).Exist(ctx)
	if err != nil || !exists {
		return ErrInvalidForeignKey
	}

	// Обновление задания
	err = s.Client.Task.UpdateOneID(taskID).
		SetInspectorID(inspectorID).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

// DeleteTask — удаление задания (только для Specialist).
func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	err := s.Client.Task.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}
