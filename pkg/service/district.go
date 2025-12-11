package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"jkh/ent"
	"jkh/ent/district"
	"jkh/pkg/models"
)

var (
	ErrDistrictNotFound = errors.New("district not found")
	ErrDistrictConflict = errors.New("district with this name already exists")
)

// DistrictService отвечает за бизнес-логику CRUD для районов
type DistrictService struct {
	Client *ent.Client
}

// Конструктор
func NewDistrictService(client *ent.Client) *DistrictService {
	return &DistrictService{Client: client}
}

// Преобразование Ent-сущности в DTO
func (s *DistrictService) toDistrictResponse(d *ent.District) *models.DistrictResponse {
	return &models.DistrictResponse{
		ID:   d.ID,
		Name: d.Name,
	}
}

// CreateDistrict — создание нового района
func (s *DistrictService) CreateDistrict(ctx context.Context, req models.CreateDistrictRequest) (*models.DistrictResponse, error) {
	d, err := s.Client.District.Create().
		SetName(req.Name).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrDistrictConflict
		}
		log.Printf("DB error creating district: %v", err)
		return nil, fmt.Errorf("database error")
	}
	return s.toDistrictResponse(d), nil
}

// ListDistricts — список всех районов
func (s *DistrictService) ListDistricts(ctx context.Context) ([]*models.DistrictResponse, error) {
	districts, err := s.Client.District.Query().All(ctx)
	if err != nil {
		log.Printf("DB error listing districts: %v", err)
		return nil, fmt.Errorf("database error")
	}

	resp := make([]*models.DistrictResponse, len(districts))
	for i, d := range districts {
		resp[i] = s.toDistrictResponse(d)
	}
	return resp, nil
}

// RetrieveDistrict — чтение района по ID
func (s *DistrictService) RetrieveDistrict(ctx context.Context, id int) (*models.DistrictResponse, error) {
	d, err := s.Client.District.Query().
		Where(district.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrDistrictNotFound
		}
		log.Printf("DB error retrieving district %d: %v", id, err)
		return nil, fmt.Errorf("database error")
	}
	return s.toDistrictResponse(d), nil
}

// UpdateDistrict — обновление района
func (s *DistrictService) UpdateDistrict(ctx context.Context, id int, req models.CreateDistrictRequest) (*models.DistrictResponse, error) {
	d, err := s.Client.District.UpdateOneID(id).
		SetName(req.Name).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrDistrictNotFound
		}
		if ent.IsConstraintError(err) {
			return nil, ErrDistrictConflict
		}
		log.Printf("DB error updating district %d: %v", id, err)
		return nil, fmt.Errorf("database error")
	}
	return s.toDistrictResponse(d), nil
}

// DeleteDistrict — удаление района
func (s *DistrictService) DeleteDistrict(ctx context.Context, id int) error {
	err := s.Client.District.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrDistrictNotFound
		}
		if ent.IsConstraintError(err) {
			return errors.New("district has active dependencies (JKH units or buildings)")
		}
		log.Printf("DB error deleting district %d: %v", id, err)
		return fmt.Errorf("database error")
	}
	return nil
}
