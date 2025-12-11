// pkg/models/user.go

package models

// CreateUserRequest — DTO для создания нового пользователя (POST)
type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Login      string `json:"login" binding:"required"`
	Password   string `json:"password" binding:"required,min=8"` 
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	// Имя роли (строка) будет преобразовано в ID в сервисном слое
	RoleName   string `json:"role_name" binding:"required,oneof=Coordinator Inspector"` 
}

// UserResponse — DTO для исходящего ответа (GET, POST)
type UserResponse struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Login      string `json:"login"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	RoleName   string `json:"role_name"`
	// Hashed password НИКОГДА не возвращается 
}

// UpdateUserRequest — DTO для обновления существующего пользователя (PUT/PATCH)
type UpdateUserRequest struct {
	// Все поля опциональны, кроме ID, который берется из URL
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	Login     *string `json:"login,omitempty"`
	Password  *string `json:"password,omitempty" binding:"omitempty,min=8"` // Только если меняем пароль
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	RoleName  *string `json:"role_name,omitempty" binding:"omitempty,oneof=Coordinator Inspector"`
}