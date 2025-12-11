// seed_admin.go

package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
	
// 	"golang.org/x/crypto/bcrypt" // Для хеширования паролей [1]
// 	"jkh/ent"
// 	"jkh/ent/role"
// 	"jkh/ent/user" // <-- Оставляем, так как используется в проверке

// 	"entgo.io/ent/dialect" // <-- ДОБАВЛЕНО
// 	_ "github.com/lib/pq"
// )

// DSN берется из client.go
//!!! Убедитесь, что эти данные совпадают с вашими данными в pkg/db/client.go
// const dsn = "host=localhost port=5432 user=postgres password=23052006 dbname=jkh-inspection sslmode=disable" 

// func main() {
// 	// 1. Подключение к БД (используя Ent без миграций)
// 	// Используем dialect.Postgres, который импортирован выше
// 	client, err := ent.Open(dialect.Postgres, dsn)
//     if err!= nil {
//         log.Fatalf("failed opening connection: %v", err)
//     }
// 	defer client.Close()

// 	ctx := context.Background()

// 	// 2. Хеширование пароля [1]
// 	const plainPassword = "AdminPassword2025!"
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
// 	if err!= nil {
// 		log.Fatalf("Failed to hash password: %v", err)
// 	}

// 	// 3. Получение ID роли 'Specialist'
// 	specialistRole, err := client.Role.Query().
// 		Where(role.NameEQ("Specialist")).
// 		Only(ctx)
// 	if err!= nil {
// 		log.Fatalf("Failed to find Specialist role. Did main.go run successfully? Error: %v", err)
// 	}

// 	// 4. Создание пользователя-специалиста
// 	_, err = client.User.Create().
// 		SetEmail("admin@gmail.com"). // Email для входа
// 		SetLogin("admin").        // Логин для входа
// 		SetPasswordHash(string(hashedPassword)).
// 		SetFirstName("Иван").
// 		SetLastName("Специалист").
// 		SetRole(specialistRole).
// 		Save(ctx)

// 	if err!= nil {
// 		// Проверяем, не является ли ошибка дублированием записи
// 		if ent.IsConstraintError(err) {
// 			// Проверяем, существует ли пользователь по email (для дополнительной ясности)
// 			if exists, _ := client.User.Query().Where(user.EmailEQ("admin@gmail.com")).Exist(ctx); exists {
// 				fmt.Println("Admin user already exists (admin@gmail.com).")
// 				return
// 			}
// 		}
// 		log.Fatalf("Failed to create admin user: %v", err)
// 	}

// 	fmt.Printf("\n--- Специалист успешно создан ---\n")
// 	fmt.Printf("Логин (Email): admin@jkh.ru\n")
// 	fmt.Printf("Пароль: %s\n", plainPassword)
// }