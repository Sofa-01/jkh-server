package models

// ============================================================================
// DTO ДЛЯ TASK (Задания на осмотр)
// ============================================================================

// CreateTaskRequest — DTO для создания задания (Coordinator).
type CreateTaskRequest struct {
    // ID здания для осмотра (обязательно).
    BuildingID int `json:"building_id" binding:"required,min=1"`
    
    // ID чек-листа для использования (обязательно).
    ChecklistID int `json:"checklist_id" binding:"required,min=1"`
    
    // ID назначенного инспектора (обязательно).
    InspectorID int `json:"inspector_id" binding:"required,min=1"`
    
    // Название задания (краткое описание).
    Title string `json:"title" binding:"required"`
    
    // Приоритет: "срочный", "высокий", "обычный", "низкий".
    Priority string `json:"priority" binding:"omitempty,oneof=срочный высокий обычный низкий"`
    
    // Подробное описание задания (опционально).
    Description *string `json:"description,omitempty"`
    
    // Планируемая дата и время осмотра (ISO 8601: "2025-04-15T10:00:00Z").
    ScheduledDate string `json:"scheduled_date" binding:"required"`
}

// TaskResponse — DTO для базового ответа (список заданий).
type TaskResponse struct {
    ID            int    `json:"id"`
    Title         string `json:"title"`
    Status        string `json:"status"`         // New, Pending, InProgress, etc.
    Priority      string `json:"priority"`
    ScheduledDate string `json:"scheduled_date"` // ISO 8601
    CreatedAt     string `json:"created_at"`
    
    // Упрощенная информация о связанных сущностях
    BuildingAddress string `json:"building_address"`
    ChecklistTitle  string `json:"checklist_title"`
    InspectorName   string `json:"inspector_name"`
}

// TaskDetailResponse — DTO для детального просмотра задания.
type TaskDetailResponse struct {
    ID            int    `json:"id"`
    Title         string `json:"title"`
    Status        string `json:"status"`
    Priority      string `json:"priority"`
    Description   string `json:"description"`
    ScheduledDate string `json:"scheduled_date"`
    CreatedAt     string `json:"created_at"`
    UpdatedAt     string `json:"updated_at"`
    
    // Детальная информация о связанных сущностях
    Building  BuildingInfo  `json:"building"`
    Checklist ChecklistInfo `json:"checklist"`
    Inspector InspectorInfo `json:"inspector"`
}

// Вспомогательные структуры для детального ответа
type BuildingInfo struct {
    ID      int    `json:"id"`
    Address string `json:"address"`
}

type ChecklistInfo struct {
    ID             int    `json:"id"`
    Title          string `json:"title"`
    InspectionType string `json:"inspection_type"`
}

type InspectorInfo struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}

// UpdateTaskStatusRequest — DTO для изменения статуса задания.
type UpdateTaskStatusRequest struct {
    Status string `json:"status" binding:"required,oneof=Pending InProgress OnReview ForRevision Approved Canceled"`
}

// AssignInspectorRequest — DTO для переназначения инспектора.
type AssignInspectorRequest struct {
    InspectorID int `json:"inspector_id" binding:"required,min=1"`
}
