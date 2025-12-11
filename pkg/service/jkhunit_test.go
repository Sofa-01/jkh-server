// pkg/service/jkhunit_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func TestJkhUnitService_CreateJkhUnit_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	// Сначала создаём район
	districtSvc := NewDistrictService(client)
	ctx := context.Background()
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Центральный"})

	svc := NewJkhUnitService(client)

	req := models.CreateJkhUnitRequest{
		Name:       "ЖЭУ-1",
		DistrictID: district.ID,
	}

	resp, err := svc.CreateJkhUnit(ctx, req)
	if err != nil {
		t.Fatalf("CreateJkhUnit failed: %v", err)
	}

	if resp.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, resp.Name)
	}
	if resp.DistrictID != district.ID {
		t.Errorf("Expected district_id %d, got %d", district.ID, resp.DistrictID)
	}
}

func TestJkhUnitService_CreateJkhUnit_InvalidDistrict(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewJkhUnitService(client)
	ctx := context.Background()

	req := models.CreateJkhUnitRequest{
		Name:       "ЖЭУ-X",
		DistrictID: 99999, // Несуществующий район
	}

	_, err := svc.CreateJkhUnit(ctx, req)
	if err != ErrDistrictFKNotFound {
		t.Errorf("Expected ErrDistrictFKNotFound, got %v", err)
	}
}

func TestJkhUnitService_CreateJkhUnit_Duplicate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	districtSvc := NewDistrictService(client)
	ctx := context.Background()
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	svc := NewJkhUnitService(client)

	req := models.CreateJkhUnitRequest{Name: "Дубликат", DistrictID: district.ID}

	_, err := svc.CreateJkhUnit(ctx, req)
	if err != nil {
		t.Fatalf("First CreateJkhUnit failed: %v", err)
	}

	_, err = svc.CreateJkhUnit(ctx, req)
	if err != ErrJkhUnitConflict {
		t.Errorf("Expected ErrJkhUnitConflict, got %v", err)
	}
}

func TestJkhUnitService_ListJkhUnits(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	districtSvc := NewDistrictService(client)
	ctx := context.Background()
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	svc := NewJkhUnitService(client)

	units := []string{"ЖЭУ-1", "ЖЭУ-2", "ЖЭУ-3"}
	for _, name := range units {
		_, err := svc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: name, DistrictID: district.ID})
		if err != nil {
			t.Fatalf("CreateJkhUnit failed: %v", err)
		}
	}

	list, err := svc.ListJkhUnits(ctx)
	if err != nil {
		t.Fatalf("ListJkhUnits failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 units, got %d", len(list))
	}
}

func TestJkhUnitService_RetrieveJkhUnit_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	districtSvc := NewDistrictService(client)
	ctx := context.Background()
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	svc := NewJkhUnitService(client)
	created, _ := svc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "Тест", DistrictID: district.ID})

	retrieved, err := svc.RetrieveJkhUnit(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveJkhUnit failed: %v", err)
	}

	if retrieved.Name != "Тест" {
		t.Errorf("Expected name 'Тест', got %s", retrieved.Name)
	}
}

func TestJkhUnitService_RetrieveJkhUnit_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewJkhUnitService(client)
	ctx := context.Background()

	_, err := svc.RetrieveJkhUnit(ctx, 99999)
	if err != ErrJkhUnitNotFound {
		t.Errorf("Expected ErrJkhUnitNotFound, got %v", err)
	}
}

func TestJkhUnitService_DeleteJkhUnit_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	districtSvc := NewDistrictService(client)
	ctx := context.Background()
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	svc := NewJkhUnitService(client)
	created, _ := svc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "Удалить", DistrictID: district.ID})

	err := svc.DeleteJkhUnit(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteJkhUnit failed: %v", err)
	}

	_, err = svc.RetrieveJkhUnit(ctx, created.ID)
	if err != ErrJkhUnitNotFound {
		t.Errorf("Expected unit to be deleted")
	}
}

