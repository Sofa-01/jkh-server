// pkg/handlers/auth_test.go

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jkh/ent"
	"jkh/pkg/models"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func setupTestClient(t *testing.T) *ent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
		db.Close()
	})

	return client
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := setupTestClient(t)
	ctx := context.Background()

	// Создаём роль
	role := client.Role.Create().SetName("Inspector").SaveX(ctx)

	// Создаём пользователя
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	client.User.Create().
		SetEmail("test@example.com").
		SetLogin("testuser").
		SetPasswordHash(string(hash)).
		SetFirstName("Test").
		SetLastName("User").
		SetRoleID(role.ID).
		SaveX(ctx)

	// Настраиваем роутер
	r := gin.New()
	authHandler := NewAuthHandler(client)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// Выполняем запрос
	reqBody := models.LoginRequest{
		Identifier: "testuser",
		Password:   "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if resp.Role != "inspector" {
		t.Errorf("Expected role 'inspector', got %s", resp.Role)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := setupTestClient(t)
	ctx := context.Background()

	role := client.Role.Create().SetName("Inspector").SaveX(ctx)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	client.User.Create().
		SetEmail("test@example.com").
		SetLogin("testuser").
		SetPasswordHash(string(hash)).
		SetFirstName("Test").
		SetLastName("User").
		SetRoleID(role.ID).
		SaveX(ctx)

	r := gin.New()
	authHandler := NewAuthHandler(client)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// Неверный пароль
	reqBody := models.LoginRequest{
		Identifier: "testuser",
		Password:   "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := setupTestClient(t)

	r := gin.New()
	authHandler := NewAuthHandler(client)
	r.POST("/api/v1/auth/login", authHandler.Login)

	reqBody := models.LoginRequest{
		Identifier: "nonexistent",
		Password:   "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := setupTestClient(t)

	r := gin.New()
	authHandler := NewAuthHandler(client)
	r.POST("/api/v1/auth/login", authHandler.Login)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
