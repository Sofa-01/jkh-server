package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"jkh/ent"
	"jkh/ent/user"
	"jkh/pkg/auth"
	"jkh/pkg/models"
)

// AuthHandler содержит Ent Client для доступа к БД
type AuthHandler struct {
	Client *ent.Client
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{Client: client}
}

// Login godoc
// @Summary      Авторизация пользователя
// @Description  Аутентификация по логину/email и паролю. Возвращает JWT access и refresh токены.
// @Tags         Авторизация
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Данные для входа (логин/email и пароль)"
// @Success      200 {object} models.LoginResponse "Успешная авторизация"
// @Failure      400 {object} map[string]string "Неверный формат запроса"
// @Failure      401 {object} map[string]string "Неверные учетные данные"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest

    // 1. Чтение и валидация JSON
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Используем контекст запроса, чтобы можно было отменить DB-операции, если клиент разорвал соединение
    ctx := c.Request.Context()

    // 2. Поиск пользователя в БД по email ИЛИ login.
    //    Сразу загружаем роль через WithRole()
    foundUser, err := h.Client.User.Query().
        Where(user.Or(
            user.EmailEQ(req.Identifier),
            user.LoginEQ(req.Identifier),
        )).
        WithRole().
        Only(ctx)

    if err != nil {
        if ent.IsNotFound(err) {
            // не раскрываем, что именно не найдено — security best practice
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }
        // internal error
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // 3. Проверка пароля (Bcrypt)
    // bcrypt.CompareHashAndPassword принимает []byte
    if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password)); err != nil {
        // неверный пароль
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Проверим, что роль подгружена
    var roleID int
    var roleName string
    if foundUser.Edges.Role != nil {
        roleID = foundUser.Edges.Role.ID
        roleName = strings.ToLower(foundUser.Edges.Role.Name)
    } else {
        // если роль отсутствует — это внутренняя ошибка данных
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User role not set"})
        return
    }

    // 4. Генерация JWT-токенов
    accessToken, refreshToken, err := auth.GenerateTokens(foundUser, roleID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
        return
    }

    // 5. Отдаём токены: вариант A — возвращаем оба в JSON
    c.JSON(http.StatusOK, models.LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        Role:         roleName,
    })

}
