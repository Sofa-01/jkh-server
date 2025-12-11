package service

import (
    "context"
    "errors"
    "fmt"
    "log"

    "jkh/ent"
    "jkh/ent/elementcatalog" // Сгенерированный Ent-пакет для работы с ElementCatalog
    "jkh/pkg/models"
)

// ============================================================================
// ОШИБКИ БИЗНЕС-ЛОГИКИ
// ============================================================================

// Централизованное определение ошибок для Service-слоя.
// Это позволяет обработчикам (handlers) различать типы ошибок через errors.Is().
var (
    // Элемент не найден в БД (404 Not Found).
    ErrElementNotFound = errors.New("element not found")
    
    // Конфликт уникальности: элемент с таким именем уже существует (409 Conflict).
    ErrElementConflict = errors.New("element name already exists")
)

// ============================================================================
// СЕРВИС
// ============================================================================

// ElementCatalogService — слой бизнес-логики для работы со справочником элементов.
// Инкапсулирует всю логику работы с БД и преобразования данных.
type ElementCatalogService struct {
    Client *ent.Client // Клиент Ent для доступа к базе данных
}

// NewElementCatalogService — конструктор сервиса.
func NewElementCatalogService(client *ent.Client) *ElementCatalogService {
    return &ElementCatalogService{Client: client}
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

// toElementResponse — преобразует Ent-сущность в DTO для ответа.
// Это обеспечивает единообразие формата данных для фронтенда.
func (s *ElementCatalogService) toElementResponse(e *ent.ElementCatalog) *models.ElementCatalogResponse {
    return &models.ElementCatalogResponse{
        ID:       e.ID,
        Name:     e.Name,
        Category: e.Category, // Ent возвращает пустую строку, если поле было NULL
    }
}

// ============================================================================
// CRUD-ОПЕРАЦИИ
// ============================================================================

// CreateElement — создание нового элемента справочника.
//
// Параметры:
//   - ctx: контекст запроса (для отмены операции при разрыве соединения)
//   - req: данные для создания элемента
//
// Возвращает:
//   - *models.ElementCatalogResponse: созданный элемент с ID
//   - error: ошибка (ErrElementConflict, если имя уже существует)
func (s *ElementCatalogService) CreateElement(ctx context.Context, req models.CreateElementCatalogRequest) (*models.ElementCatalogResponse, error) {
    // Инициализация билдера для создания записи в БД
    create := s.Client.ElementCatalog.Create().
        SetName(req.Name) // Обязательное поле

    // Если категория передана (не nil), устанавливаем её
    if req.Category != nil {
        create.SetCategory(*req.Category)
    }
    // Если req.Category == nil, Ent установит значение по умолчанию (пустая строка)

    // Выполнение запроса к БД
    e, err := create.Save(ctx)
    if err != nil {
        // Проверка на ошибку уникальности (UNIQUE constraint violation)
        if ent.IsConstraintError(err) {
            return nil, ErrElementConflict
        }
        // Логируем внутреннюю ошибку БД для отладки
        log.Printf("DB error creating element: %v", err)
        return nil, fmt.Errorf("database error")
    }

    // Преобразуем Ent-сущность в DTO и возвращаем
    return s.toElementResponse(e), nil
}

// ListElements — получение списка всех элементов справочника.
//
// Возвращает:
//   - []*models.ElementCatalogResponse: массив всех элементов
//   - error: ошибка БД (если произошла)
func (s *ElementCatalogService) ListElements(ctx context.Context) ([]*models.ElementCatalogResponse, error) {
    // Запрос всех записей из таблицы element_catalogs
    elements, err := s.Client.ElementCatalog.Query().All(ctx)
    if err != nil {
        return nil, fmt.Errorf("database error")
    }

    // Преобразуем каждую Ent-сущность в DTO
    resp := make([]*models.ElementCatalogResponse, len(elements))
    for i, e := range elements {
        resp[i] = s.toElementResponse(e)
    }

    return resp, nil
}

// RetrieveElement — получение одного элемента по ID.
//
// Параметры:
//   - ctx: контекст запроса
//   - id: ID элемента
//
// Возвращает:
//   - *models.ElementCatalogResponse: найденный элемент
//   - error: ErrElementNotFound, если элемент не найден
func (s *ElementCatalogService) RetrieveElement(ctx context.Context, id int) (*models.ElementCatalogResponse, error) {
    // Запрос элемента по ID
    e, err := s.Client.ElementCatalog.Query().
        Where(elementcatalog.IDEQ(id)). // Фильтр WHERE id = ?
        Only(ctx)                        // Ожидаем ровно одну запись (или ошибку)
    
    if err != nil {
        // Если запись не найдена, возвращаем специфичную ошибку
        if ent.IsNotFound(err) {
            return nil, ErrElementNotFound
        }
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    return s.toElementResponse(e), nil
}

// UpdateElement — обновление существующего элемента.
//
// Параметры:
//   - ctx: контекст запроса
//   - id: ID обновляемого элемента
//   - req: новые данные
//
// Возвращает:
//   - *models.ElementCatalogResponse: обновленный элемент
//   - error: ErrElementNotFound (если не найден) или ErrElementConflict (если имя занято)
func (s *ElementCatalogService) UpdateElement(ctx context.Context, id int, req models.CreateElementCatalogRequest) (*models.ElementCatalogResponse, error) {
    // Инициализация билдера для обновления записи по ID
    update := s.Client.ElementCatalog.UpdateOneID(id).
        SetName(req.Name) // Обновляем имя

    // Обработка nullable-поля category
    if req.Category != nil {
        update.SetCategory(*req.Category) // Устанавливаем новое значение
    } else {
        update.ClearCategory() // Очищаем поле (устанавливаем пустую строку)
    }

    // Выполнение запроса
    e, err := update.Save(ctx)
    if err != nil {
        // Проверка, что запись с таким ID существует
        if ent.IsNotFound(err) {
            return nil, ErrElementNotFound
        }
        // Проверка на конфликт уникальности (новое имя уже занято)
        if ent.IsConstraintError(err) {
            return nil, ErrElementConflict
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    return s.toElementResponse(e), nil
}

// DeleteElement — удаление элемента из справочника.
//
// Параметры:
//   - ctx: контекст запроса
//   - id: ID удаляемого элемента
//
// Возвращает:
//   - error: ErrElementNotFound (если не найден) или ошибку, если элемент используется
func (s *ElementCatalogService) DeleteElement(ctx context.Context, id int) error {
    // Попытка удалить элемент по ID
    err := s.Client.ElementCatalog.DeleteOneID(id).Exec(ctx)
    if err != nil {
        // Элемент не найден
        if ent.IsNotFound(err) {
            return ErrElementNotFound
        }
        // Элемент используется в чек-листах (FK constraint violation)
        if ent.IsConstraintError(err) {
            return errors.New("element has active dependencies (used in checklists)")
        }
        return fmt.Errorf("database error: %w", err)
    }
    
    return nil
}
