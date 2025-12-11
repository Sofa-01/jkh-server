//pkg/service/jkhunit.go

package service

import (
	"context"
	"errors"
	"fmt"

	"jkh/ent"
	"jkh/ent/district"
	"jkh/ent/jkhunit"
	"jkh/pkg/models"
)

// Ошибки, которые возвращает сервис
var (
	ErrJkhUnitNotFound    = errors.New("jkh unit not found")        // ЖЭУ не найден
	ErrJkhUnitConflict    = errors.New("jkh unit name already exists") // Дублирующееся имя
	ErrDistrictFKNotFound = errors.New("specified district not found") // Район не найден
)

// JkhUnitService содержит клиент Ent для работы с таблицей JkhUnit
type JkhUnitService struct {
	Client *ent.Client
}

// Конструктор
func NewJkhUnitService(client *ent.Client) *JkhUnitService {
	return &JkhUnitService{Client: client}
}

// toJkhUnitResponse — преобразование Ent-сущности в DTO
func (s *JkhUnitService) toJkhUnitResponse(j *ent.JkhUnit) *models.JkhUnitResponse {
	districtName := ""
	if j.Edges.District != nil {
		districtName = j.Edges.District.Name
	}

	return &models.JkhUnitResponse{
		ID:           j.ID,
		Name:         j.Name,
		DistrictID:   j.DistrictID,
		DistrictName: districtName,
	}
}

// CreateJkhUnit — создаем ЖЭУ с проверкой существования района
func (s *JkhUnitService) CreateJkhUnit(ctx context.Context, req models.CreateJkhUnitRequest) (*models.JkhUnitResponse, error) {

	// Проверка существования FK
	exists, err := s.Client.District.Query().Where(district.IDEQ(req.DistrictID)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return nil, ErrDistrictFKNotFound
	}

	// Создаем ЖЭУ
	j, err := s.Client.JkhUnit.Create().
		SetName(req.Name).
		SetDistrictID(req.DistrictID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrJkhUnitConflict
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Загружаем с District для DTO
	j, err = s.Client.JkhUnit.Query().
		Where(jkhunit.IDEQ(j.ID)).
		WithDistrict().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to reload created jkh unit: %w", err)
	}

	return s.toJkhUnitResponse(j), nil
}

// ListJkhUnits — получение списка всех ЖЭУ
func (s *JkhUnitService) ListJkhUnits(ctx context.Context) ([]*models.JkhUnitResponse, error) {
	jkhUnits, err := s.Client.JkhUnit.Query().
		WithDistrict(). // для DTO нужен DistrictName
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Преобразуем в DTO
	resp := make([]*models.JkhUnitResponse, len(jkhUnits))
	for i, j := range jkhUnits {
		resp[i] = s.toJkhUnitResponse(j)
	}

	return resp, nil
}

// RetrieveJkhUnit — чтение ЖЭУ по ID
func (s *JkhUnitService) RetrieveJkhUnit(ctx context.Context, id int) (*models.JkhUnitResponse, error) {
	j, err := s.Client.JkhUnit.Query().
		Where(jkhunit.IDEQ(id)).
		WithDistrict().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrJkhUnitNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return s.toJkhUnitResponse(j), nil
}

// UpdateJkhUnit — обновление ЖЭУ
func (s *JkhUnitService) UpdateJkhUnit(ctx context.Context, id int, req models.CreateJkhUnitRequest) (*models.JkhUnitResponse, error) {
	// Проверка существования нового DistrictID
	exists, err := s.Client.District.Query().Where(district.IDEQ(req.DistrictID)).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return nil, ErrDistrictFKNotFound
	}

	j, err := s.Client.JkhUnit.UpdateOneID(id).
		SetName(req.Name).
		SetDistrictID(req.DistrictID).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrJkhUnitNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrJkhUnitConflict
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Загружаем для DTO
	j, err = s.Client.JkhUnit.Query().
		Where(jkhunit.IDEQ(j.ID)).
		WithDistrict().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to reload updated jkh unit: %w", err)
	}

	return s.toJkhUnitResponse(j), nil
}

// DeleteJkhUnit — удаление ЖЭУ
func (s *JkhUnitService) DeleteJkhUnit(ctx context.Context, id int) error {
	err := s.Client.JkhUnit.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrJkhUnitNotFound
		}
		if ent.IsConstraintError(err) {
			return errors.New("jkh unit has active dependencies") // FK Constraint
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}
