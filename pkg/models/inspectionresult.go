package models

// ============================================================================
// DTO ДЛЯ INSPECTIONRESULT (Результаты осмотра)
// ============================================================================

// CreateInspectionResultRequest — DTO для создания/обновления результата проверки элемента.
type CreateInspectionResultRequest struct {
	// ID элемента чек-листа (из ChecklistElement).
	ChecklistElementID int `json:"checklist_element_id" binding:"required,min=1"`
	
	// Статус состояния элемента.
	// Допустимые значения: "Исправное", "Удовлетворительное", "Неудовлетворительное", "Аварийное"
	ConditionStatus string `json:"condition_status" binding:"required,oneof=Исправное Удовлетворительное Неудовлетворительное Аварийное"`
	
	// Комментарий инспектора (опционально).
	Comment *string `json:"comment,omitempty"`
}

// InspectionResultResponse — DTO для исходящих ответов.
type InspectionResultResponse struct {
	TaskID             int    `json:"task_id"`
	ChecklistElementID int    `json:"checklist_element_id"`
	
	// Информация об элементе
	ElementName     string `json:"element_name"`     // Название элемента (например, "Кровля")
	ElementCategory string `json:"element_category"` // Категория
	OrderIndex      int    `json:"order_index"`      // Порядок в чек-листе
	
	// Результат проверки
	ConditionStatus string `json:"condition_status"`
	Comment         string `json:"comment"`
	
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TaskResultsSummary — DTO для сводки по заданию (все результаты + прогресс).
type TaskResultsSummary struct {
	TaskID            int                         `json:"task_id"`
	TaskTitle         string                      `json:"task_title"`
	TotalElements     int                         `json:"total_elements"`      // Всего элементов в чек-листе
	CompletedElements int                         `json:"completed_elements"`  // Заполнено результатов
	Results           []InspectionResultResponse  `json:"results"`             // Список результатов
}
