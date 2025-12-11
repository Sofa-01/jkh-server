// pkg/service/elementcatalog_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func ptr(s string) *string {
	return &s
}

func TestElementCatalogService_CreateElement_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	req := models.CreateElementCatalogRequest{
		Name:     "Фундамент",
		Category: ptr("Несущие конструкции"),
	}

	resp, err := svc.CreateElement(ctx, req)
	if err != nil {
		t.Fatalf("CreateElement failed: %v", err)
	}

	if resp.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, resp.Name)
	}
	if resp.Category != "Несущие конструкции" {
		t.Errorf("Expected category 'Несущие конструкции', got %s", resp.Category)
	}
}

func TestElementCatalogService_CreateElement_Duplicate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	req := models.CreateElementCatalogRequest{Name: "Кровля", Category: ptr("Покрытия")}

	_, err := svc.CreateElement(ctx, req)
	if err != nil {
		t.Fatalf("First CreateElement failed: %v", err)
	}

	_, err = svc.CreateElement(ctx, req)
	if err != ErrElementConflict {
		t.Errorf("Expected ErrElementConflict, got %v", err)
	}
}

func TestElementCatalogService_ListElements(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	elements := []models.CreateElementCatalogRequest{
		{Name: "Стены", Category: ptr("Ограждающие")},
		{Name: "Окна", Category: ptr("Светопрозрачные")},
		{Name: "Двери", Category: ptr("Проёмы")},
	}

	for _, e := range elements {
		_, err := svc.CreateElement(ctx, e)
		if err != nil {
			t.Fatalf("CreateElement failed: %v", err)
		}
	}

	list, err := svc.ListElements(ctx)
	if err != nil {
		t.Fatalf("ListElements failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(list))
	}
}

func TestElementCatalogService_RetrieveElement_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	created, _ := svc.CreateElement(ctx, models.CreateElementCatalogRequest{Name: "Лестница"})

	retrieved, err := svc.RetrieveElement(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveElement failed: %v", err)
	}

	if retrieved.Name != "Лестница" {
		t.Errorf("Expected name 'Лестница', got %s", retrieved.Name)
	}
}

func TestElementCatalogService_RetrieveElement_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	_, err := svc.RetrieveElement(ctx, 99999)
	if err != ErrElementNotFound {
		t.Errorf("Expected ErrElementNotFound, got %v", err)
	}
}

func TestElementCatalogService_UpdateElement_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	created, _ := svc.CreateElement(ctx, models.CreateElementCatalogRequest{Name: "Старое"})

	updated, err := svc.UpdateElement(ctx, created.ID, models.CreateElementCatalogRequest{Name: "Новое", Category: ptr("Категория")})
	if err != nil {
		t.Fatalf("UpdateElement failed: %v", err)
	}

	if updated.Name != "Новое" {
		t.Errorf("Expected name 'Новое', got %s", updated.Name)
	}
}

func TestElementCatalogService_DeleteElement_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	created, _ := svc.CreateElement(ctx, models.CreateElementCatalogRequest{Name: "Для удаления"})

	err := svc.DeleteElement(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteElement failed: %v", err)
	}

	_, err = svc.RetrieveElement(ctx, created.ID)
	if err != ErrElementNotFound {
		t.Errorf("Expected element to be deleted")
	}
}

func TestElementCatalogService_DeleteElement_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewElementCatalogService(client)
	ctx := context.Background()

	err := svc.DeleteElement(ctx, 99999)
	if err != ErrElementNotFound {
		t.Errorf("Expected ErrElementNotFound, got %v", err)
	}
}
