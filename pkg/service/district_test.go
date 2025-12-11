// pkg/service/district_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func TestDistrictService_CreateDistrict_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	req := models.CreateDistrictRequest{
		Name: "Центральный район",
	}

	resp, err := svc.CreateDistrict(ctx, req)
	if err != nil {
		t.Fatalf("CreateDistrict failed: %v", err)
	}

	if resp.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, resp.Name)
	}
	if resp.ID == 0 {
		t.Error("Expected non-zero ID")
	}
}

func TestDistrictService_CreateDistrict_Duplicate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	req := models.CreateDistrictRequest{Name: "Дубликат"}

	// Первый — успешно
	_, err := svc.CreateDistrict(ctx, req)
	if err != nil {
		t.Fatalf("First CreateDistrict failed: %v", err)
	}

	// Второй с тем же именем — ошибка
	_, err = svc.CreateDistrict(ctx, req)
	if err != ErrDistrictConflict {
		t.Errorf("Expected ErrDistrictConflict, got %v", err)
	}
}

func TestDistrictService_ListDistricts(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	// Создаём несколько районов
	districts := []string{"Северный", "Южный", "Западный"}
	for _, name := range districts {
		_, err := svc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: name})
		if err != nil {
			t.Fatalf("CreateDistrict failed: %v", err)
		}
	}

	// Получаем список
	list, err := svc.ListDistricts(ctx)
	if err != nil {
		t.Fatalf("ListDistricts failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 districts, got %d", len(list))
	}
}

func TestDistrictService_RetrieveDistrict_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	created, err := svc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Тестовый"})
	if err != nil {
		t.Fatalf("CreateDistrict failed: %v", err)
	}

	retrieved, err := svc.RetrieveDistrict(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveDistrict failed: %v", err)
	}

	if retrieved.Name != "Тестовый" {
		t.Errorf("Expected name 'Тестовый', got %s", retrieved.Name)
	}
}

func TestDistrictService_RetrieveDistrict_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	_, err := svc.RetrieveDistrict(ctx, 99999)
	if err != ErrDistrictNotFound {
		t.Errorf("Expected ErrDistrictNotFound, got %v", err)
	}
}

func TestDistrictService_UpdateDistrict_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	created, _ := svc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Старое имя"})

	updated, err := svc.UpdateDistrict(ctx, created.ID, models.CreateDistrictRequest{Name: "Новое имя"})
	if err != nil {
		t.Fatalf("UpdateDistrict failed: %v", err)
	}

	if updated.Name != "Новое имя" {
		t.Errorf("Expected name 'Новое имя', got %s", updated.Name)
	}
}

func TestDistrictService_UpdateDistrict_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	_, err := svc.UpdateDistrict(ctx, 99999, models.CreateDistrictRequest{Name: "Тест"})
	if err != ErrDistrictNotFound {
		t.Errorf("Expected ErrDistrictNotFound, got %v", err)
	}
}

func TestDistrictService_DeleteDistrict_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	created, _ := svc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Для удаления"})

	err := svc.DeleteDistrict(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteDistrict failed: %v", err)
	}

	// Проверяем что удалён
	_, err = svc.RetrieveDistrict(ctx, created.ID)
	if err != ErrDistrictNotFound {
		t.Errorf("Expected district to be deleted")
	}
}

func TestDistrictService_DeleteDistrict_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewDistrictService(client)
	ctx := context.Background()

	err := svc.DeleteDistrict(ctx, 99999)
	if err != ErrDistrictNotFound {
		t.Errorf("Expected ErrDistrictNotFound, got %v", err)
	}
}

