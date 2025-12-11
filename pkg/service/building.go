package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"jkh/ent"
	"jkh/ent/building"
	"jkh/ent/district"
	"jkh/ent/jkhunit"
	"jkh/ent/user"
	"jkh/pkg/models"
)

// Общие ошибки бизнес-логики
var (
	ErrBuildingNotFound = errors.New("building not found")
	ErrBuildingConflict = errors.New("building address already exists")
	ErrFKNotFound       = errors.New("one or more foreign keys not found (District, JKH Unit, or Inspector)")
)

// BuildingService — слой бизнес-логики.
type BuildingService struct {
	Client *ent.Client
}

func NewBuildingService(client *ent.Client) *BuildingService {
	return &BuildingService{Client: client}
}

// toBuildingResponse — преобразование Ent → DTO.
// Gemini добавлял это как вспомогательную функцию.
func (s *BuildingService) toBuildingResponse(b *ent.Building) *models.BuildingResponse {
	resp := &models.BuildingResponse{
		ID:               b.ID,
		Address:          b.Address,
		ConstructionYear: b.ConstructionYear,
		Description:      b.Description,
		PhotoPath:        b.Photo,
	}

	// Добавляем имена FK. Работает только если было WithDistrict / WithJkhUnit / WithInspector.
	if b.Edges.District != nil {
		resp.DistrictName = b.Edges.District.Name
	}
	if b.Edges.JkhUnit != nil {
		resp.JkhUnitName = b.Edges.JkhUnit.Name
	}
	if b.Edges.Inspector != nil {
		resp.InspectorName = fmt.Sprintf("%s %s",
			b.Edges.Inspector.FirstName,
			b.Edges.Inspector.LastName)
	}

	return resp
}

// checkFKs — это обеспечивает ссылочную целостность.
// Я добавил обработку ошибок.
func (s *BuildingService) checkFKs(ctx context.Context, districtID, jkhUnitID int, inspectorID *int) error {
	// Проверка District
	dExists, err := s.Client.District.Query().Where(district.IDEQ(districtID)).Exist(ctx)
	if err != nil {
		return err
	}

	// Проверка JKH Unit
	jExists, err := s.Client.JkhUnit.Query().Where(jkhunit.IDEQ(jkhUnitID)).Exist(ctx)
	if err != nil {
		return err
	}

	if !dExists || !jExists {
		return ErrFKNotFound
	}

	// Проверка Inspector (nullable)
	if inspectorID != nil {
		iExists, err := s.Client.User.Query().Where(user.IDEQ(*inspectorID)).Exist(ctx)
		if err != nil {
			return err
		}
		if !iExists {
			return ErrFKNotFound
		}
	}

	return nil
}

// CreateBuilding — создание объекта.
func (s *BuildingService) CreateBuilding(ctx context.Context, req models.CreateBuildingRequest) (*models.BuildingResponse, error) {
	// Проверка FK
	if err := s.checkFKs(ctx, req.DistrictID, req.JkhUnitID, req.InspectorID); err != nil {
		return nil, err
	}

	// Создание сущности
	create := s.Client.Building.Create().
		SetAddress(req.Address).
		SetConstructionYear(req.ConstructionYear).
		SetDistrictID(req.DistrictID).
		SetJkhUnitID(req.JkhUnitID)

	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.Photo != nil {
		create.SetPhoto(*req.Photo)
	}
	if req.InspectorID != nil {
		create.SetInspectorID(*req.InspectorID)
	}

	b, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrBuildingConflict
		}
		log.Printf("DB error creating building: %v", err)
		return nil, fmt.Errorf("database error")
	}

	// Догружаем связи
	b, err = s.Client.Building.Query().
		Where(building.IDEQ(b.ID)).
		WithDistrict().
		WithJkhUnit().
		WithInspector().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created building: %w", err)
	}

	return s.toBuildingResponse(b), nil
}

// ListBuildings — исправлено полностью
func (s *BuildingService) ListBuildings(ctx context.Context) ([]*models.BuildingResponse, error) {
	buildings, err := s.Client.Building.Query().
		WithDistrict().
		WithJkhUnit().
		WithInspector().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error")
	}

	resp := make([]*models.BuildingResponse, len(buildings))
	for i, b := range buildings {
		resp[i] = s.toBuildingResponse(b)
	}

	return resp, nil
}

// RetrieveBuilding — получить по ID.
func (s *BuildingService) RetrieveBuilding(ctx context.Context, id int) (*models.BuildingResponse, error) {
	b, err := s.Client.Building.Query().
		Where(building.IDEQ(id)).
		WithDistrict().
		WithJkhUnit().
		WithInspector().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrBuildingNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return s.toBuildingResponse(b), nil
}

// UpdateBuilding — обновление.
func (s *BuildingService) UpdateBuilding(ctx context.Context, id int, req models.CreateBuildingRequest) (*models.BuildingResponse, error) {
	if err := s.checkFKs(ctx, req.DistrictID, req.JkhUnitID, req.InspectorID); err != nil {
		return nil, err
	}

	update := s.Client.Building.UpdateOneID(id).
		SetAddress(req.Address).
		SetConstructionYear(req.ConstructionYear).
		SetDistrictID(req.DistrictID).
		SetJkhUnitID(req.JkhUnitID)

	if req.Description != nil {
		update.SetDescription(*req.Description)
	} else {
		update.ClearDescription()
	}
	if req.Photo != nil {
		update.SetPhoto(*req.Photo)
	} else {
		update.ClearPhoto()
	}
	if req.InspectorID != nil {
		update.SetInspectorID(*req.InspectorID)
	} else {
		update.ClearInspector()
	}

	b, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrBuildingNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrBuildingConflict
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	b, err = s.Client.Building.Query().
		Where(building.IDEQ(b.ID)).
		WithDistrict().
		WithJkhUnit().
		WithInspector().
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated building: %w", err)
	}

	return s.toBuildingResponse(b), nil
}

// DeleteBuilding — удаление.
func (s *BuildingService) DeleteBuilding(ctx context.Context, id int) error {
	err := s.Client.Building.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrBuildingNotFound
		}
		if ent.IsConstraintError(err) {
			// Объект привязан к активным заданиям
			return errors.New("building has active dependencies (tasks)")
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}
