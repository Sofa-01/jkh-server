// pkg/service/user.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"jkh/ent"
	"jkh/ent/role"
	"jkh/ent/user"
	"jkh/pkg/models"
)

// Определение доменных ошибок
var (
	ErrRoleNotFound = errors.New("role not found")
	ErrUserConflict = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// UserService отвечает за бизнес-логику CRUD для пользователей
type UserService struct {
	Client *ent.Client
}

func NewUserService(client *ent.Client) *UserService {
	return &UserService{Client: client}
}

// hashPassword хеширует чистый пароль с помощью Bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// findRoleID находит ID роли по ее строковому имени
func (s *UserService) findRoleID(ctx context.Context, roleName string) (int, error) {
	r, err := s.Client.Role.Query().
		Where(role.NameEQ(roleName)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, ErrRoleNotFound
		}
		return 0, err
	}
	return r.ID, nil
}

// CreateUser - создает нового пользователя (Use Case)
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	// 1. Найти ID роли
	roleID, err := s.findRoleID(ctx, req.RoleName)
	if err != nil {
		return nil, err
	}

	// 2. Хеширование пароля
	hashedPwd, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 3. Создание пользователя в БД
	u, err := s.Client.User.Create().
		SetEmail(req.Email).
		SetLogin(req.Login).
		SetPasswordHash(hashedPwd).
		SetFirstName(req.FirstName).
		SetLastName(req.LastName).
		SetRoleID(roleID). // Запись FK
		Save(ctx)

	if err != nil {
		// Обработка ошибки уникальности (email, login)
		if ent.IsConstraintError(err) {
			return nil, ErrUserConflict
		}
		log.Printf("DB error creating user: %v", err)
		return nil, fmt.Errorf("database error")
	}

	// 4. Подгрузим роль, чтобы корректно вернуть RoleName в ответе
	uWithRole, err := s.Client.User.Query().
		Where(user.IDEQ(u.ID)).
		WithRole().
		Only(ctx)
	if err != nil {
		// Не фатально — вернём минимальный ответ, но логируем
		log.Printf("Warning: created user but failed to load role edge: %v", err)
		return s.toUserResponse(u), nil
	}

	return s.toUserResponse(uWithRole), nil
}

// toUserResponse - вспомогательная функция для преобразования Ent-сущности в DTO (безопасность)
func (s *UserService) toUserResponse(u *ent.User) *models.UserResponse {
	// Получаем имя роли через edge, чтобы не делать лишний запрос
	roleName := "unknown"
	if u.Edges.Role != nil {
		roleName = u.Edges.Role.Name
	}

	return &models.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Login:     u.Login,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		RoleName:  roleName,
	}
}

// ListUsers - получает список всех пользователей
func (s *UserService) ListUsers(ctx context.Context) ([]*models.UserResponse, error) {
	// Загружаем пользователей, сразу присоединяя роль (для получения RoleName)
	users, err := s.Client.User.Query().
		WithRole().
		All(ctx)
	if err != nil {
		log.Printf("DB error listing users: %v", err)
		return nil, fmt.Errorf("database error")
	}

	resp := make([]*models.UserResponse, len(users))
	for i, u := range users {
		resp[i] = s.toUserResponse(u)
	}
	return resp, nil
}

// findUserAndRoleByUserID - вспомогательная функция для получения пользователя и его роли
// Этот метод инкапсулирует логику Ent
func (s *UserService) findUserAndRoleByUserID(ctx context.Context, id int) (*ent.User, error) {
	u, err := s.Client.User.Query().
		Where(user.IDEQ(id)).
		WithRole(). // Присоединяем роль, чтобы получить RoleName для DTO
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) { // Обработка ошибки Ent: не найдено
			return nil, ErrUserNotFound
		}
		// Обработка общей ошибки БД
		log.Printf("DB error finding user %d: %v", id, err)
		return nil, fmt.Errorf("database error")
	}
	return u, nil
}

// RetrieveUser - чтение пользователя по ID (Use Case)
func (s *UserService) RetrieveUser(ctx context.Context, id int) (*models.UserResponse, error) {
	u, err := s.findUserAndRoleByUserID(ctx, id)
	if err != nil {
		// Ошибка уже преобразована в ErrUserNotFound в findUserAndRoleByUserID
		return nil, err
	}
	// Преобразуем Ent-сущность в безопасный DTO (скрывая хеш пароля)
	return s.toUserResponse(u), nil
}

// UpdateUser - обновляет существующего пользователя
func (s *UserService) UpdateUser(ctx context.Context, targetUserID int, authenticatedUserID int, req models.UpdateUserRequest) (*models.UserResponse, error) {
    if targetUserID == authenticatedUserID {
        if req.RoleName != nil {
            return nil, errors.New("cannot change own role")
        }
    }

    tx, err := s.Client.Tx(ctx)
    if err != nil {
        return nil, fmt.Errorf("starting transaction: %w", err)
    }
    defer func() {
        if err != nil {
            tx.Rollback()
            return
        }
        tx.Commit()
    }()

    update := tx.User.UpdateOneID(targetUserID)

    if req.Email != nil {
        update.SetEmail(*req.Email)
    }
    if req.Login != nil {
        update.SetLogin(*req.Login)
    }
    if req.FirstName != nil {
        update.SetFirstName(*req.FirstName)
    }
    if req.LastName != nil {
        update.SetLastName(*req.LastName)
    }

    if req.Password != nil && len(*req.Password) > 0 {
        hashedPwd, err := hashPassword(*req.Password)
        if err != nil {
            return nil, fmt.Errorf("password hashing failed: %w", err)
        }
        update.SetPasswordHash(hashedPwd)
    }

    if req.RoleName != nil {
        roleID, err := s.findRoleID(ctx, *req.RoleName)
        if err != nil {
            return nil, err
        }
        update.SetRoleID(roleID)
    }

    u, err := update.Save(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, ErrUserNotFound
        }
        if ent.IsConstraintError(err) {
            return nil, ErrUserConflict
        }
        log.Printf("DB error updating user %d: %v", targetUserID, err)
        return nil, fmt.Errorf("database error")
    }

    u, err = s.findUserAndRoleByUserID(ctx, u.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch updated user: %w", err)
    }

    return s.toUserResponse(u), nil
}

// DeleteUser - удаляет пользователя (Hard Delete с проверкой зависимостей)
func (s *UserService) DeleteUser(ctx context.Context, targetUserID int, authenticatedUserID int) error {
    // 1. Проверка запрета на самоудаление
    if targetUserID == authenticatedUserID {
        return errors.New("cannot delete own account")
    }

    // 2. Попытка удаления
    err := s.Client.User.DeleteOneID(targetUserID).Exec(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return ErrUserNotFound
        }
        if ent.IsConstraintError(err) {
            return errors.New("user has active dependencies (tasks, buildings, etc.)")
        }
        log.Printf("DB error deleting user %d: %v", targetUserID, err)
        return fmt.Errorf("database error")
    }

    return nil
}