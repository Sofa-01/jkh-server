package models

// ============================================================================
// DTO ДЛЯ CHECKLIST (Чек-листы)
// ============================================================================

// CreateChecklistRequest — DTO для создания/обновления чек-листа.
type CreateChecklistRequest struct {
    // Название чек-листа (например, "Весенний осмотр многоквартирных домов").
    // Поле обязательно и уникально в БД.
    Title string `json:"title" binding:"required"`
    
    // Тип осмотра: "spring" (весенний), "winter" (зимний), "partial" (частичный).
    // Значение по умолчанию в БД: "partial".
    InspectionType string `json:"inspection_type" binding:"required,oneof=spring winter partial"`
    
    // Описание чек-листа (опционально).
    Description *string `json:"description,omitempty"`
}

// ChecklistResponse — DTO для исходящих ответов (базовая информация о чек-листе).
type ChecklistResponse struct {
    ID             int    `json:"id"`
    Title          string `json:"title"`
    InspectionType string `json:"inspection_type"` // spring/winter/partial
    Description    string `json:"description"`
    CreatedAt      string `json:"created_at"` // ISO 8601 формат
}

// ChecklistDetailResponse — DTO для детального ответа (чек-лист + список элементов).
// Используется при GET /checklists/:id для отображения всех элементов с порядком.
type ChecklistDetailResponse struct {
    ID             int                      `json:"id"`
    Title          string                   `json:"title"`
    InspectionType string                   `json:"inspection_type"`
    Description    string                   `json:"description"`
    CreatedAt      string                   `json:"created_at"`
    Elements       []ChecklistElementDetail `json:"elements"` // Список элементов в чек-листе
}

// ChecklistElementDetail — информация об элементе в чек-листе.
type ChecklistElementDetail struct {
    ElementID   int    `json:"element_id"`   // ID элемента из ElementCatalog
    ElementName string `json:"element_name"` // Название элемента (например, "Кровля")
    Category    string `json:"category"`     // Категория элемента
    OrderIndex  int    `json:"order_index"`  // Порядок проверки (1, 2, 3...)
}

// ============================================================================
// DTO ДЛЯ CHECKLISTELEMENT (Управление элементами в чек-листе)
// ============================================================================

// AddElementToChecklistRequest — DTO для добавления элемента в чек-лист.
type AddElementToChecklistRequest struct {
    // ID элемента из справочника ElementCatalog.
    ElementID int `json:"element_id" binding:"required,min=1"`
    
    // Порядок проверки элемента в чек-листе (опционально).
    // Если не указан, элемент добавляется в конец списка.
    OrderIndex *int `json:"order_index,omitempty"`
}

// UpdateElementOrderRequest — DTO для изменения порядка элемента в чек-листе.
type UpdateElementOrderRequest struct {
    // Новый порядок проверки элемента.
    OrderIndex int `json:"order_index" binding:"required,min=1"`
}
