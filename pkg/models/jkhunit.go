package models

// CreateJkhUnitRequest — DTO для создания или обновления ЖЭУ
type CreateJkhUnitRequest struct {
	Name       string `json:"name" binding:"required"`
	DistrictID int    `json:"district_id" binding:"required,min=1"`
}

// JkhUnitResponse — DTO ответа
type JkhUnitResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DistrictID   int    `json:"district_id"`
	DistrictName string `json:"district_name"`
}
