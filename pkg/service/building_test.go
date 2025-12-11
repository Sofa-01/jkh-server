// pkg/service/building_test.go

package service

import (
	"context"
	"testing"

	"jkh/pkg/models"
	"jkh/pkg/testutil"
)

func TestBuildingService_CreateBuilding_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	// Создаём район и ЖЭУ
	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ-1", DistrictID: district.ID})

	// Создаём здание
	svc := NewBuildingService(client)
	req := models.CreateBuildingRequest{
		Address:          "ул. Тестовая, д. 1",
		DistrictID:       district.ID,
		JkhUnitID:        jkhUnit.ID,
		ConstructionYear: 1980,
	}

	resp, err := svc.CreateBuilding(ctx, req)
	if err != nil {
		t.Fatalf("CreateBuilding failed: %v", err)
	}

	if resp.Address != req.Address {
		t.Errorf("Expected address %s, got %s", req.Address, resp.Address)
	}
	if resp.ConstructionYear != 1980 {
		t.Errorf("Expected construction year 1980, got %d", resp.ConstructionYear)
	}
}

func TestBuildingService_CreateBuilding_InvalidFK(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewBuildingService(client)
	ctx := context.Background()

	req := models.CreateBuildingRequest{
		Address:    "ул. Тестовая, д. 1",
		DistrictID: 99999,
		JkhUnitID:  99999,
	}

	_, err := svc.CreateBuilding(ctx, req)
	if err != ErrFKNotFound {
		t.Errorf("Expected ErrFKNotFound, got %v", err)
	}
}

func TestBuildingService_CreateBuilding_Duplicate(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ", DistrictID: district.ID})

	svc := NewBuildingService(client)

	req := models.CreateBuildingRequest{
		Address:    "Дубликат",
		DistrictID: district.ID,
		JkhUnitID:  jkhUnit.ID,
	}

	_, err := svc.CreateBuilding(ctx, req)
	if err != nil {
		t.Fatalf("First CreateBuilding failed: %v", err)
	}

	_, err = svc.CreateBuilding(ctx, req)
	if err != ErrBuildingConflict {
		t.Errorf("Expected ErrBuildingConflict, got %v", err)
	}
}

func TestBuildingService_ListBuildings(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ", DistrictID: district.ID})

	svc := NewBuildingService(client)

	addresses := []string{"ул. Первая, 1", "ул. Вторая, 2", "ул. Третья, 3"}
	for _, addr := range addresses {
		_, err := svc.CreateBuilding(ctx, models.CreateBuildingRequest{
			Address:    addr,
			DistrictID: district.ID,
			JkhUnitID:  jkhUnit.ID,
		})
		if err != nil {
			t.Fatalf("CreateBuilding failed: %v", err)
		}
	}

	list, err := svc.ListBuildings(ctx)
	if err != nil {
		t.Fatalf("ListBuildings failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 buildings, got %d", len(list))
	}
}

func TestBuildingService_RetrieveBuilding_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ", DistrictID: district.ID})

	svc := NewBuildingService(client)
	created, _ := svc.CreateBuilding(ctx, models.CreateBuildingRequest{
		Address:    "Тест",
		DistrictID: district.ID,
		JkhUnitID:  jkhUnit.ID,
	})

	retrieved, err := svc.RetrieveBuilding(ctx, created.ID)
	if err != nil {
		t.Fatalf("RetrieveBuilding failed: %v", err)
	}

	if retrieved.Address != "Тест" {
		t.Errorf("Expected address 'Тест', got %s", retrieved.Address)
	}
}

func TestBuildingService_RetrieveBuilding_NotFound(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	svc := NewBuildingService(client)
	ctx := context.Background()

	_, err := svc.RetrieveBuilding(ctx, 99999)
	if err != ErrBuildingNotFound {
		t.Errorf("Expected ErrBuildingNotFound, got %v", err)
	}
}

func TestBuildingService_UpdateBuilding_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ", DistrictID: district.ID})

	svc := NewBuildingService(client)
	created, _ := svc.CreateBuilding(ctx, models.CreateBuildingRequest{
		Address:    "Старый",
		DistrictID: district.ID,
		JkhUnitID:  jkhUnit.ID,
	})

	updated, err := svc.UpdateBuilding(ctx, created.ID, models.CreateBuildingRequest{
		Address:    "Новый",
		DistrictID: district.ID,
		JkhUnitID:  jkhUnit.ID,
	})
	if err != nil {
		t.Fatalf("UpdateBuilding failed: %v", err)
	}

	if updated.Address != "Новый" {
		t.Errorf("Expected address 'Новый', got %s", updated.Address)
	}
}

func TestBuildingService_DeleteBuilding_Success(t *testing.T) {
	client := testutil.SetupTestDB(t)
	defer client.Close()

	ctx := context.Background()

	districtSvc := NewDistrictService(client)
	district, _ := districtSvc.CreateDistrict(ctx, models.CreateDistrictRequest{Name: "Район"})

	jkhSvc := NewJkhUnitService(client)
	jkhUnit, _ := jkhSvc.CreateJkhUnit(ctx, models.CreateJkhUnitRequest{Name: "ЖЭУ", DistrictID: district.ID})

	svc := NewBuildingService(client)
	created, _ := svc.CreateBuilding(ctx, models.CreateBuildingRequest{
		Address:    "Удалить",
		DistrictID: district.ID,
		JkhUnitID:  jkhUnit.ID,
	})

	err := svc.DeleteBuilding(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteBuilding failed: %v", err)
	}

	_, err = svc.RetrieveBuilding(ctx, created.ID)
	if err != ErrBuildingNotFound {
		t.Errorf("Expected building to be deleted")
	}
}
