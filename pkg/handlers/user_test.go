// pkg/handlers/user_test.go

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"jkh/ent"
	"jkh/pkg/models"
	"jkh/pkg/service"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func setupUserTest(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

	// Создаём роли
	for _, roleName := range []string{"Specialist", "Coordinator", "Inspector"} {
		client.Role.Create().SetName(roleName).SaveX(ctx)
	}

	t.Cleanup(func() {
		client.Close()
		db.Close()
	})

	userService := service.NewUserService(client)
	userHandler := NewUserHandler(userService)

	r := gin.New()
	r.POST("/api/v1/users", userHandler.CreateUser)
	r.GET("/api/v1/users", userHandler.ListUsers)
	r.GET("/api/v1/users/:id", userHandler.GetUser)
	r.PUT("/api/v1/users/:id", userHandler.UpdateUser)
	r.DELETE("/api/v1/users/:id", userHandler.DeleteUser)

	return r, client
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	r, _ := setupUserTest(t)

	reqBody := models.CreateUserRequest{
		Email:     "test@example.com",
		Login:     "testuser",
		Password:  "password123",
		FirstName: "Иван",
		LastName:  "Иванов",
		RoleName:  "Inspector",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", resp.Email)
	}
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	r, _ := setupUserTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	r, client := setupUserTest(t)

	ctx := context.Background()
	role, _ := client.Role.Query().First(ctx)

	// Создаём пользователей
	client.User.Create().
		SetEmail("user1@test.com").SetLogin("user1").SetPasswordHash("hash").
		SetFirstName("A").SetLastName("A").SetRoleID(role.ID).SaveX(ctx)
	client.User.Create().
		SetEmail("user2@test.com").SetLogin("user2").SetPasswordHash("hash").
		SetFirstName("B").SetLastName("B").SetRoleID(role.ID).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp []models.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 users, got %d", len(resp))
	}
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	r, client := setupUserTest(t)

	ctx := context.Background()
	role, _ := client.Role.Query().First(ctx)
	user := client.User.Create().
		SetEmail("get@test.com").SetLogin("getuser").SetPasswordHash("hash").
		SetFirstName("Get").SetLastName("User").SetRoleID(role.ID).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%d", user.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	r, _ := setupUserTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	r, _ := setupUserTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}


