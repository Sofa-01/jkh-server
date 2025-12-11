// pkg/service/checklist_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func TestChecklistService_CreateChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	desc := "Описание чеклиста"
	req := models.CreateChecklistRequest{
		Title:          "Типовой чеклист",
		InspectionType: "spring",
		Description:    &desc,
	}

	resp, err := svc.CreateChecklist(ctx, req)
	if err != nil {
		t.Fatalf("CreateChecklist failed: %v", err)
	}

	if resp.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, resp.Title)
	}
	if resp.InspectionType != req.InspectionType {
		t.Errorf("Expected inspection_type %s, got %s", req.InspectionType, resp.InspectionType)
	}
}

func TestChecklistService_CreateChecklist_Duplicate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	req := models.CreateChecklistRequest{Title: "Дубликат", InspectionType: "winter"}

	_, err := svc.CreateChecklist(ctx, req)
	if err != nil {
		t.Fatalf("First CreateChecklist failed: %v", err)
	}

	_, err = svc.CreateChecklist(ctx, req)
	if err != ErrChecklistConflict {
		t.Errorf("Expected ErrChecklistConflict, got %v", err)
	}
}

func TestChecklistService_ListChecklists(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	names := []string{"Чеклист 1", "Чеклист 2", "Чеклист 3"}
	for _, name := range names {
		_, err := svc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: name, InspectionType: "partial"})
		if err != nil {
			t.Fatalf("CreateChecklist failed: %v", err)
		}
	}

	list, err := svc.ListChecklists(ctx)
	if err != nil {
		t.Fatalf("ListChecklists failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 checklists, got %d", len(list))
	}
}

func TestChecklistService_RetrieveChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	created, _ := svc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: "Тест", InspectionType: "spring"})

	retrieved, err := svc.RetrieveChecklist(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveChecklist failed: %v", err)
	}

	if retrieved.Title != "Тест" {
		t.Errorf("Expected title 'Тест', got %s", retrieved.Title)
	}
}

func TestChecklistService_RetrieveChecklist_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	_, err := svc.RetrieveChecklist(ctx, 99999)
	if err != ErrChecklistNotFound {
		t.Errorf("Expected ErrChecklistNotFound, got %v", err)
	}
}

func TestChecklistService_UpdateChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	created, _ := svc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: "Старое", InspectionType: "spring"})

	updated, err := svc.UpdateChecklist(ctx, created.ID, models.CreateChecklistRequest{Title: "Новое", InspectionType: "winter"})
	if err != nil {
		t.Fatalf("UpdateChecklist failed: %v", err)
	}

	if updated.Title != "Новое" {
		t.Errorf("Expected title 'Новое', got %s", updated.Title)
	}
}

func TestChecklistService_DeleteChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	created, _ := svc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: "Удалить", InspectionType: "partial"})

	err := svc.DeleteChecklist(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteChecklist failed: %v", err)
	}

	_, err = svc.RetrieveChecklist(ctx, created.ID)
	if err != ErrChecklistNotFound {
		t.Errorf("Expected checklist to be deleted")
	}
}

func TestChecklistService_DeleteChecklist_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewChecklistService(client)
	ctx := context.Background()

	err := svc.DeleteChecklist(ctx, 99999)
	if err != ErrChecklistNotFound {
		t.Errorf("Expected ErrChecklistNotFound, got %v", err)
	}
}

func TestChecklistService_AddElementToChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	// Создаём элемент в справочнике
	elemSvc := NewElementCatalogService(client)
	elem, _ := elemSvc.CreateElement(ctx, models.CreateElementCatalogRequest{Name: "Фундамент"})

	// Создаём чеклист
	checklistSvc := NewChecklistService(client)
	checklist, _ := checklistSvc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: "Тест", InspectionType: "spring"})

	// Добавляем элемент
	orderIdx := 1
	err := checklistSvc.AddElementToChecklist(ctx, checklist.ID, models.AddElementToChecklistRequest{
		ElementID:  elem.ID,
		OrderIndex: &orderIdx,
	})
	if err != nil {
		t.Fatalf("AddElementToChecklist failed: %v", err)
	}

	// Проверяем что элемент добавлен
	retrieved, _ := checklistSvc.RetrieveChecklist(ctx, checklist.ID)
	if len(retrieved.Elements) != 1 {
		t.Errorf("Expected 1 element, got %d", len(retrieved.Elements))
	}
}

func TestChecklistService_AddElementToChecklist_ChecklistNotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	checklistSvc := NewChecklistService(client)

	orderIdx := 1
	err := checklistSvc.AddElementToChecklist(ctx, 99999, models.AddElementToChecklistRequest{
		ElementID:  1,
		OrderIndex: &orderIdx,
	})
	if err != ErrChecklistNotFound {
		t.Errorf("Expected ErrChecklistNotFound, got %v", err)
	}
}

func TestChecklistService_RemoveElementFromChecklist_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	// Создаём элемент
	elemSvc := NewElementCatalogService(client)
	elem, _ := elemSvc.CreateElement(ctx, models.CreateElementCatalogRequest{Name: "Элемент"})

	// Создаём чеклист и добавляем элемент
	checklistSvc := NewChecklistService(client)
	checklist, _ := checklistSvc.CreateChecklist(ctx, models.CreateChecklistRequest{Title: "Тест", InspectionType: "partial"})

	orderIdx := 1
	checklistSvc.AddElementToChecklist(ctx, checklist.ID, models.AddElementToChecklistRequest{
		ElementID:  elem.ID,
		OrderIndex: &orderIdx,
	})

	// Получаем ID добавленного ChecklistElement
	retrieved, _ := checklistSvc.RetrieveChecklist(ctx, checklist.ID)
	elemID := retrieved.Elements[0].ElementID

	// Удаляем
	err := checklistSvc.RemoveElementFromChecklist(ctx, checklist.ID, elemID)
	if err != nil {
		t.Fatalf("RemoveElementFromChecklist failed: %v", err)
	}

	// Проверяем что удалён
	retrieved, _ = checklistSvc.RetrieveChecklist(ctx, checklist.ID)
	if len(retrieved.Elements) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(retrieved.Elements))
	}
}
