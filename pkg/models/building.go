package models

// CreateBuildingRequest — DTO для создания/обновления Объекта (PUT/POST).
// Используется в хендлере. Позволяет не дублировать две разные модели.
type CreateBuildingRequest struct {
	Address          string  `json:"address" binding:"required"`
	ConstructionYear int     `json:"construction_year"`
	Description      *string `json:"description,omitempty"`   // nullable
	Photo            *string `json:"photo_path,omitempty"`    // nullable
	
	// Обязательные внешние ключи
	DistrictID       int     `json:"district_id" binding:"required,min=1"`
	JkhUnitID        int     `json:"jkh_unit_id" binding:"required,min=1"`
	
	// FK (optional) — назначение инспектора необязательно
	InspectorID      *int    `json:"inspector_id,omitempty"`
}

// BuildingResponse — DTO для исходящих ответов.
// Форматируется под потребности фронтенда.
type BuildingResponse struct {
	ID               int       `json:"id"`
	Address          string    `json:"address"`
	ConstructionYear int       `json:"construction_year"`
	Description      string    `json:"description"`
	PhotoPath        string    `json:"photo_path"`

	// Имена связанных сущностей
	DistrictName     string    `json:"district_name"`
	JkhUnitName      string    `json:"jkh_unit_name"`
	InspectorName    string    `json:"inspector_name,omitempty"`
}
