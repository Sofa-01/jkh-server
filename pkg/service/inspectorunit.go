// pkg/service/inspectorunit.go

package service

import (
	"context"
	"errors"
	"fmt"

	"jkh/ent"
	"jkh/ent/inspectorunit"
	"jkh/ent/jkhunit"
	"jkh/ent/user"
	"jkh/pkg/models"
)

var (
	ErrInspectorAssignmentExists   = errors.New("inspector already assigned to this jkh unit")
	ErrInspectorAssignmentNotFound = errors.New("inspector assignment not found")
)

type InspectorUnitService struct {
	Client *ent.Client
}

func NewInspectorUnitService(client *ent.Client) *InspectorUnitService {
	return &InspectorUnitService{Client: client}
}

// AssignInspector — назначить инспектора на ЖЭУ
func (s *InspectorUnitService) AssignInspector(ctx context.Context, jkhUnitID, inspectorID int) error {
	// Проверки существования FK
	exists, err := s.Client.User.Query().Where(user.IDEQ(inspectorID)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	exists, err = s.Client.JkhUnit.Query().Where(jkhunit.IDEQ(jkhUnitID)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if !exists {
		return ErrJkhUnitNotFound
	}

	// Проверка дубликата
	dup, err := s.Client.InspectorUnit.Query().Where(
		inspectorunit.UserIDEQ(inspectorID),
		inspectorunit.JkhUnitIDEQ(jkhUnitID),
	).Exist(ctx)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if dup {
		return ErrInspectorAssignmentExists
	}

	_, err = s.Client.InspectorUnit.Create().
		SetUserID(inspectorID).
		SetJkhUnitID(jkhUnitID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create inspector assignment: %w", err)
	}
	return nil
}

// UnassignInspector — удалить назначение по паре (jkhUnitID, inspectorID)
func (s *InspectorUnitService) UnassignInspector(ctx context.Context, jkhUnitID, inspectorID int) error {
	// Попытка удаления по условиям
	res, err := s.Client.InspectorUnit.Delete().
		Where(
			inspectorunit.UserIDEQ(inspectorID),
			inspectorunit.JkhUnitIDEQ(jkhUnitID),
		).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrInspectorAssignmentNotFound
		}
		return fmt.Errorf("database error: %w", err)
	}
	if res == 0 {
		return ErrInspectorAssignmentNotFound
	}
	return nil
}

// ListInspectorsForUnit — получить список пользователей (Inspector) назначенных на ЖЭУ
func (s *InspectorUnitService) ListInspectorsForUnit(ctx context.Context, jkhUnitID int) ([]*models.UserResponse, error) {
	users, err := s.Client.User.Query().
		Where(user.HasAssignedUnitsWith(inspectorunit.JkhUnitIDEQ(jkhUnitID))).
		WithRole().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	resp := make([]*models.UserResponse, len(users))
	for i, u := range users {
		roleName := "unknown"
		if u.Edges.Role != nil {
			roleName = u.Edges.Role.Name
		}
		resp[i] = &models.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Login:     u.Login,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			RoleName:  roleName,
		}
	}
	return resp, nil
}

// ListUnitsForInspector — получить список ЖЭУ для заданного инспектора
func (s *InspectorUnitService) ListUnitsForInspector(ctx context.Context, inspectorID int) ([]*models.JkhUnitResponse, error) {
	units, err := s.Client.JkhUnit.Query().
		Where(jkhunit.HasAssignedInspectorsWith(inspectorunit.UserIDEQ(inspectorID))).
		WithDistrict().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	resp := make([]*models.JkhUnitResponse, len(units))
	for i, j := range units {
		districtName := ""
		if j.Edges.District != nil {
			districtName = j.Edges.District.Name
		}
		resp[i] = &models.JkhUnitResponse{
			ID:           j.ID,
			Name:         j.Name,
			DistrictID:   j.DistrictID,
			DistrictName: districtName,
		}
	}
	return resp, nil
}
