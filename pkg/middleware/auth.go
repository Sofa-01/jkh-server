// pkg/middleware/auth.go

package middleware

import (
	"fmt"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"jkh/pkg/auth"
)

// Роли для удобства (Role ID: 1-Specialist, 2-Coordinator, 3-Inspector)
const (
    RoleSpecialist  = 1
    RoleCoordinator = 2
    RoleInspector   = 3
)

var jwtSecret = []byte("YOUR_ULTRA_SECURE_SECRET_KEY_12345") // Должен совпадать с ключом в pkg/auth/jwt.go

// AuthRequired проверяет наличие и валидность Access Token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Извлекаем токен без "Bearer "
		tokenString = tokenString[7:]

		claims := &auth.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Сохраняем UserID и RoleID в контексте Gin для дальнейшего использования
		c.Set("userID", claims.UserID)
		c.Set("roleID", claims.RoleID)
		
		c.Next() // Передаем управление следующему обработчику
	}
}

// RBACMiddleware проверяет, соответствует ли роль пользователя требуемому уровню доступа
func RBACMiddleware(requiredRoleID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, exists := c.Get("roleID")
		if !exists {
			// Ошибка: токен не был проверен (AuthRequired пропущен?)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User role not authenticated"})
			return
		}

		userRoleID, ok := roleID.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Invalid role format"})
			return
		}
		
		// Логика RBAC: Чем МЕНЬШЕ число, тем ВЫШЕ привилегии (1-Specialist, 3-Inspector)
		// Если роль пользователя (userRoleID) НИЖЕ, чем требуемая (requiredRoleID), то доступ разрешён.
		// Пример: Specialist (1) -> Coordinator (2) = OK
		// Пример: Inspector (3) -> Specialist (1) = DENIED
		
		if userRoleID > requiredRoleID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: insufficient privileges"})
			return
		}

		c.Next()
	}
}
