// pkg/service/user_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Login:     "testuser",
		Password:  "password123",
		FirstName: "Иван",
		LastName:  "Иванов",
		RoleName:  "Inspector",
	}

	resp, err := svc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if resp.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.Email)
	}
	if resp.Login != req.Login {
		t.Errorf("Expected login %s, got %s", req.Login, resp.Login)
	}
	if resp.FirstName != req.FirstName {
		t.Errorf("Expected first_name %s, got %s", req.FirstName, resp.FirstName)
	}
	if resp.RoleName != "Inspector" {
		t.Errorf("Expected role Inspector, got %s", resp.RoleName)
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	req := models.CreateUserRequest{
		Email:     "duplicate@example.com",
		Login:     "user1",
		Password:  "password123",
		FirstName: "Пётр",
		LastName:  "Петров",
		RoleName:  "Inspector",
	}

	// Первый пользователь — успешно
	_, err := svc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("First CreateUser failed: %v", err)
	}

	// Второй пользователь с тем же email — должна быть ошибка
	req.Login = "user2" // меняем логин
	_, err = svc.CreateUser(ctx, req)
	if err != ErrUserConflict {
		t.Errorf("Expected ErrUserConflict, got %v", err)
	}
}

func TestUserService_CreateUser_InvalidRole(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	req := models.CreateUserRequest{
		Email:     "test@example.com",
		Login:     "testuser",
		Password:  "password123",
		FirstName: "Иван",
		LastName:  "Иванов",
		RoleName:  "NonExistentRole",
	}

	_, err := svc.CreateUser(ctx, req)
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestUserService_ListUsers(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	// Создаём несколько пользователей
	users := []models.CreateUserRequest{
		{Email: "user1@test.com", Login: "user1", Password: "pass", FirstName: "A", LastName: "A", RoleName: "Inspector"},
		{Email: "user2@test.com", Login: "user2", Password: "pass", FirstName: "B", LastName: "B", RoleName: "Coordinator"},
	}

	for _, u := range users {
		_, err := svc.CreateUser(ctx, u)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	}

	// Получаем список
	list, err := svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 users, got %d", len(list))
	}
}

func TestUserService_RetrieveUser_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	req := models.CreateUserRequest{
		Email:     "retrieve@test.com",
		Login:     "retrieveuser",
		Password:  "password123",
		FirstName: "Тест",
		LastName:  "Тестов",
		RoleName:  "Inspector",
	}

	created, err := svc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Получаем по ID
	retrieved, err := svc.RetrieveUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveUser failed: %v", err)
	}

	if retrieved.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, retrieved.Email)
	}
}

func TestUserService_RetrieveUser_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	_, err := svc.RetrieveUser(ctx, 99999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	// Создаём двух пользователей
	admin, _ := svc.CreateUser(ctx, models.CreateUserRequest{
		Email: "admin@test.com", Login: "admin", Password: "pass",
		FirstName: "Admin", LastName: "Admin", RoleName: "Specialist",
	})

	user, _ := svc.CreateUser(ctx, models.CreateUserRequest{
		Email: "user@test.com", Login: "user", Password: "pass",
		FirstName: "User", LastName: "User", RoleName: "Inspector",
	})

	// Admin удаляет user
	err := svc.DeleteUser(ctx, user.ID, admin.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Проверяем что пользователь удалён
	_, err = svc.RetrieveUser(ctx, user.ID)
	if err != ErrUserNotFound {
		t.Errorf("Expected user to be deleted, but got: %v", err)
	}
}

func TestUserService_DeleteUser_CannotDeleteSelf(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewUserService(client)
	ctx := context.Background()

	user, _ := svc.CreateUser(ctx, models.CreateUserRequest{
		Email: "self@test.com", Login: "self", Password: "pass",
		FirstName: "Self", LastName: "Self", RoleName: "Specialist",
	})

	// Пользователь пытается удалить себя
	err := svc.DeleteUser(ctx, user.ID, user.ID)
	if err == nil {
		t.Error("Expected error when deleting self, got nil")
	}
}

