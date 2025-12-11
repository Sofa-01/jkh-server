package auth

import (
	"time"
    "jkh/ent"
    "github.com/golang-jwt/jwt/v5"
)

// Ключ для подписи токенов
var jwtSecret = []byte("YOUR_ULTRA_SECURE_SECRET_KEY_12345")

// UserClaims — содержит данные о пользователе и его роль (Role ID)
type UserClaims struct {
    UserID int `json:"user_id"`
    RoleID int `json:"role_id"`
    jwt.RegisteredClaims
}

// GenerateTokens создает AT и RT, используя наш секретный ключ.
func GenerateTokens(user *ent.User, roleID int) (string, string, error) {
	// Access Token (короткий срок жизни, 30 минут)
	accessClaims := &UserClaims{
		UserID: user.ID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)), 
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	
	// Refresh Token (длинный срок жизни)
	refreshClaims := &UserClaims{
		UserID: user.ID,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), 
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Подписываем токены
	at, err := accessToken.SignedString(jwtSecret)
	if err!= nil {
		return "", "", err
	}
	rt, err := refreshToken.SignedString(jwtSecret)
	if err!= nil {
		return "", "", err
	}

	return at, rt, nil
}