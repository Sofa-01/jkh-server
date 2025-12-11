package models

// CreateDistrictRequest — DTO для создания или обновления района
type CreateDistrictRequest struct {
	Name string `json:"name" binding:"required"` // Название района (обязательное поле)
}

// DistrictResponse — DTO для исходящего ответа
type DistrictResponse struct {
	ID   int    `json:"id"`   // Уникальный идентификатор района
	Name string `json:"name"` // Название района
}
