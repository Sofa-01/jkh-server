// pkg/handlers/user.go

package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

// UserHandler содержит ссылку на UserService
type UserHandler struct {
	Service *service.UserService
}

// Конструктор
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{Service: s}
}

// CreateUser godoc
// @Summary      Создать пользователя
// @Description  Создание нового пользователя в системе (инспектор, координатор или специалист)
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateUserRequest true "Данные нового пользователя"
// @Success      201 {object} models.UserResponse "Пользователь успешно создан"
// @Failure      400 {object} map[string]string "Неверный запрос или роль не найдена"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      409 {object} map[string]string "Пользователь с таким email/login уже существует"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest

	// 1. Валидация JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	// 2. Вызов сервисного слоя
	resp, err := h.Service.CreateUser(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrRoleNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role name specified"})
			return
		case service.ErrUserConflict:
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email or login already exists"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// 3. Ответ
	c.JSON(http.StatusCreated, resp)
}

// ListUsers godoc
// @Summary      Получить список пользователей
// @Description  Возвращает список всех пользователей системы
// @Tags         Пользователи
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.UserResponse "Список пользователей"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.Service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user list"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// parseID извлекает и парсит ID из параметра URI
func parseID(c *gin.Context) (int, error) {
	// Получаем строку ID из URL (например, "123" из /users/123)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr) // Преобразуем строку в число
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}
	return id, nil
}

// GetUser godoc
// @Summary      Получить пользователя по ID
// @Description  Возвращает информацию о конкретном пользователе
// @Tags         Пользователи
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID пользователя"
// @Success      200 {object} models.UserResponse "Данные пользователя"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	resp, err := h.Service.RetrieveUser(c.Request.Context(), id)
	if err != nil {
		// Обработка доменных ошибок (ErrUserNotFound)
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	// Успех: HTTP 200 OK
	c.JSON(http.StatusOK, resp)
}

// UpdateUser godoc
// @Summary      Обновить пользователя
// @Description  Обновление данных пользователя (email, имя, роль и т.д.)
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID пользователя"
// @Param        request body models.UpdateUserRequest true "Данные для обновления"
// @Success      200 {object} models.UserResponse "Обновленные данные пользователя"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Нельзя изменить свою роль"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Failure      409 {object} map[string]string "Email или Login уже заняты"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var req models.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
        return
    }

    authUserID, _ := c.Get("userID")

    resp, err := h.Service.UpdateUser(c.Request.Context(), id, authUserID.(int), req)
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        if errors.Is(err, service.ErrUserConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "Email or Login already exists"})
            return
        }
        if strings.Contains(err.Error(), "cannot change own role") {
            c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify own role"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

// DeleteUser godoc
// @Summary      Удалить пользователя
// @Description  Удаление пользователя из системы
// @Tags         Пользователи
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID пользователя"
// @Success      204 "Пользователь успешно удален"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      403 {object} map[string]string "Нельзя удалить свой аккаунт"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Failure      409 {object} map[string]string "У пользователя есть активные зависимости"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
    id, err := parseID(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    authUserID, _ := c.Get("userID")

    err = h.Service.DeleteUser(c.Request.Context(), id, authUserID.(int))
    if err != nil {
        if errors.Is(err, service.ErrUserNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        if strings.Contains(err.Error(), "cannot delete own account") {
            c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete own account"})
            return
        }
        if strings.Contains(err.Error(), "active dependencies") {
            c.JSON(http.StatusConflict, gin.H{"error": "User has active dependencies (tasks, buildings, etc.) and cannot be deleted"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}



