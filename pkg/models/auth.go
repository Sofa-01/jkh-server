package models

type LoginRequest struct{
	Identifier string `json:"identifier" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct{
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Role string `json:"role"` // Роль для фронтенда (specialist, coordinator, inspector)
}