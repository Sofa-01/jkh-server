package service

import (
    "context"
    "errors"
    "fmt"
    "log"

    "jkh/ent"
    "jkh/ent/checklist"
    "jkh/ent/checklistelement"
    "jkh/ent/elementcatalog"
    "jkh/pkg/models"
)

// ============================================================================
// ОШИБКИ БИЗНЕС-ЛОГИКИ
// ============================================================================

var (
    // Чек-лист не найден (404 Not Found).
    ErrChecklistNotFound = errors.New("checklist not found")
    
    // Конфликт уникальности: чек-лист с таким названием уже существует (409 Conflict).
    ErrChecklistConflict = errors.New("checklist title already exists")
    
    // Элемент уже добавлен в этот чек-лист (409 Conflict).
    ErrElementAlreadyInChecklist = errors.New("element already added to this checklist")
    
    // Связь checklist-element не найдена (404 Not Found).
    ErrChecklistElementNotFound = errors.New("element not found in this checklist")
)

// ============================================================================
// СЕРВИС
// ============================================================================

// ChecklistService — слой бизнес-логики для работы с чек-листами.
type ChecklistService struct {
    Client *ent.Client
}

func NewChecklistService(client *ent.Client) *ChecklistService {
    return &ChecklistService{Client: client}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// toChecklistResponse — преобразует Ent-сущность в базовый DTO.
func (s *ChecklistService) toChecklistResponse(c *ent.Checklist) *models.ChecklistResponse {
    return &models.ChecklistResponse{
        ID:             c.ID,
        Title:          c.Title,
        InspectionType: string(c.InspectionType), // Enum → string
        Description:    c.Description,
        CreatedAt:      c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), // ISO 8601
    }
}

// toChecklistDetailResponse — преобразует чек-лист + элементы в детальный DTO.
func (s *ChecklistService) toChecklistDetailResponse(c *ent.Checklist) *models.ChecklistDetailResponse {
    resp := &models.ChecklistDetailResponse{
        ID:             c.ID,
        Title:          c.Title,
        InspectionType: string(c.InspectionType),
        Description:    c.Description,
        CreatedAt:      c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        Elements:       []models.ChecklistElementDetail{},
    }

    // Преобразуем связи ChecklistElement → ElementCatalog в DTO
    if c.Edges.Elements != nil {
        for _, ce := range c.Edges.Elements {
            elem := models.ChecklistElementDetail{
                ElementID:  ce.ElementID,
                OrderIndex: ce.OrderIndex,
            }
            
            // Если загружен ElementCatalog (через WithElementCatalog()), добавляем его данные
            if ce.Edges.ElementCatalog != nil {
                elem.ElementName = ce.Edges.ElementCatalog.Name
                elem.Category = ce.Edges.ElementCatalog.Category
            }
            
            resp.Elements = append(resp.Elements, elem)
        }
    }

    return resp
}

// ============================================================================
// CRUD ДЛЯ CHECKLIST
// ============================================================================

// CreateChecklist — создание нового чек-листа.
func (s *ChecklistService) CreateChecklist(ctx context.Context, req models.CreateChecklistRequest) (*models.ChecklistResponse, error) {
    // Инициализация билдера
    create := s.Client.Checklist.Create().
        SetTitle(req.Title).
        SetInspectionType(checklist.InspectionType(req.InspectionType)) // string → Enum

    // Установка опционального описания
    if req.Description != nil {
        create.SetDescription(*req.Description)
    }

    // Сохранение в БД
    c, err := create.Save(ctx)
    if err != nil {
        if ent.IsConstraintError(err) {
            return nil, ErrChecklistConflict // Название уже существует
        }
        log.Printf("DB error creating checklist: %v", err)
        return nil, fmt.Errorf("database error")
    }

    return s.toChecklistResponse(c), nil
}

// ListChecklists — получение списка всех чек-листов.
func (s *ChecklistService) ListChecklists(ctx context.Context) ([]*models.ChecklistResponse, error) {
    checklists, err := s.Client.Checklist.Query().All(ctx)
    if err != nil {
        return nil, fmt.Errorf("database error")
    }

    resp := make([]*models.ChecklistResponse, len(checklists))
    for i, c := range checklists {
        resp[i] = s.toChecklistResponse(c)
    }

    return resp, nil
}

// RetrieveChecklist — получение детальной информации о чек-листе (со списком элементов).
func (s *ChecklistService) RetrieveChecklist(ctx context.Context, id int) (*models.ChecklistDetailResponse, error) {
    // Запрос чек-листа с загрузкой связанных элементов и их данных из ElementCatalog
    c, err := s.Client.Checklist.Query().
        Where(checklist.IDEQ(id)).
        WithElements(func(q *ent.ChecklistElementQuery) {
            // Загружаем данные из ElementCatalog для каждого элемента
            q.WithElementCatalog().
                Order(ent.Asc(checklistelement.FieldOrderIndex)) // Сортировка по порядку
        }).
        Only(ctx)

    if err != nil {
        if ent.IsNotFound(err) {
            return nil, ErrChecklistNotFound
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return s.toChecklistDetailResponse(c), nil
}

// UpdateChecklist — обновление чек-листа.
func (s *ChecklistService) UpdateChecklist(ctx context.Context, id int, req models.CreateChecklistRequest) (*models.ChecklistResponse, error) {
    update := s.Client.Checklist.UpdateOneID(id).
        SetTitle(req.Title).
        SetInspectionType(checklist.InspectionType(req.InspectionType))

    if req.Description != nil {
        update.SetDescription(*req.Description)
    } else {
        update.ClearDescription()
    }

    c, err := update.Save(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, ErrChecklistNotFound
        }
        if ent.IsConstraintError(err) {
            return nil, ErrChecklistConflict
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return s.toChecklistResponse(c), nil
}

// DeleteChecklist — удаление чек-листа.
func (s *ChecklistService) DeleteChecklist(ctx context.Context, id int) error {
    err := s.Client.Checklist.DeleteOneID(id).Exec(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return ErrChecklistNotFound
        }
        if ent.IsConstraintError(err) {
            return errors.New("checklist has active dependencies (used in tasks)")
        }
        return fmt.Errorf("database error: %w", err)
    }
    return nil
}

// ============================================================================
// УПРАВЛЕНИЕ ЭЛЕМЕНТАМИ В ЧЕК-ЛИСТЕ (M:M через ChecklistElement)
// ============================================================================

// AddElementToChecklist — добавление элемента в чек-лист.
func (s *ChecklistService) AddElementToChecklist(ctx context.Context, checklistID int, req models.AddElementToChecklistRequest) error {
    // 1. Проверка существования чек-листа
    exists, err := s.Client.Checklist.Query().Where(checklist.IDEQ(checklistID)).Exist(ctx)
    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }
    if !exists {
        return ErrChecklistNotFound
    }

    // 2. Проверка существования элемента в справочнике
    elemExists, err := s.Client.ElementCatalog.Query().Where(elementcatalog.IDEQ(req.ElementID)).Exist(ctx)
    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }
    if !elemExists {
        return ErrElementNotFound // Используем ошибку из elementcatalog.go
    }

    // 3. Проверка, что элемент еще не добавлен в этот чек-лист
    alreadyAdded, err := s.Client.ChecklistElement.Query().
        Where(
            checklistelement.ChecklistIDEQ(checklistID),
            checklistelement.ElementIDEQ(req.ElementID),
        ).
        Exist(ctx)
    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }
    if alreadyAdded {
        return ErrElementAlreadyInChecklist
    }

    // 4. Определение order_index
    orderIndex := 1 // По умолчанию
    if req.OrderIndex != nil {
        orderIndex = *req.OrderIndex
    } else {
        // Если не указан, добавляем элемент в конец списка
        maxOrder, err := s.Client.ChecklistElement.Query().
            Where(checklistelement.ChecklistIDEQ(checklistID)).
            Aggregate(ent.Max(checklistelement.FieldOrderIndex)).
            Int(ctx)
        if err == nil && maxOrder > 0 {
            orderIndex = maxOrder + 1
        }
    }

    // 5. Создание записи в ChecklistElement
    _, err = s.Client.ChecklistElement.Create().
        SetChecklistID(checklistID).
        SetElementID(req.ElementID).
        SetOrderIndex(orderIndex).
        Save(ctx)

    if err != nil {
        if ent.IsConstraintError(err) {
            return ErrElementAlreadyInChecklist
        }
        log.Printf("DB error adding element to checklist: %v", err)
        return fmt.Errorf("database error")
    }

    return nil
}

// RemoveElementFromChecklist — удаление элемента из чек-листа.
func (s *ChecklistService) RemoveElementFromChecklist(ctx context.Context, checklistID, elementID int) error {
    // Удаление записи из ChecklistElement по композитному ключу
    deleted, err := s.Client.ChecklistElement.Delete().
        Where(
            checklistelement.ChecklistIDEQ(checklistID),
            checklistelement.ElementIDEQ(elementID),
        ).
        Exec(ctx)

    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }

    if deleted == 0 {
        return ErrChecklistElementNotFound // Связь не найдена
    }

    return nil
}

// UpdateElementOrder — изменение порядка элемента в чек-листе.
func (s *ChecklistService) UpdateElementOrder(ctx context.Context, checklistID, elementID, newOrder int) error {
    // Обновление order_index для конкретной записи ChecklistElement
    updated, err := s.Client.ChecklistElement.Update().
        Where(
            checklistelement.ChecklistIDEQ(checklistID),
            checklistelement.ElementIDEQ(elementID),
        ).
        SetOrderIndex(newOrder).
        Save(ctx)

    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }

    if updated == 0 {
        return ErrChecklistElementNotFound
    }

    return nil
}
