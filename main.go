package main

import (
	"context"
	"fmt"
	"log"

	"jkh/ent"
	"jkh/ent/role" // Импортируем модель для работы с ролями
	"jkh/pkg/db"
	"jkh/pkg/server"

	_ "jkh/docs" // Swagger документация (сгенерированная)
)

// @title           JKH Inspection API
// @version         1.0
// @description     REST API для системы осмотра жилых помещений ЖКХ.
// @description     Система предназначена для координации работы инспекторов,
// @description     управления заданиями на осмотр зданий и формирования отчётности.

// @contact.name   Техническая поддержка
// @contact.email  support@jkh.local

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите JWT токен в формате: Bearer {token}

func main() {
	log.Println("Starting JKH Inspection Backend...")
	
	// 1. Инициализация клиента БД Ent и выполнение миграций
	entClient := db.NewClient()
	defer func() {
		if err := entClient.Close(); err!= nil {
			log.Printf("Error closing DB client: %v", err)
		}
	}()

	// 2. Добавление базовых ролей (Specialist, Coordinator, Inspector)
	seedDatabase(entClient)

	// 3. Инициализация и запуск HTTP-сервера Gin
	r := server.SetupRouter(entClient)
	
	log.Fatal(r.Run(":8080")) // Сервер будет запущен на порту 8080
}

// seedDatabase создает необходимые базовые данные (роли)
func seedDatabase(client *ent.Client) {
    ctx := context.Background()

    roles := []string{"Specialist", "Coordinator", "Inspector"}

    for _, roleName := range roles {
        // Проверяем, существует ли роль
        count, err := client.Role.Query().
            Where(role.NameEQ(roleName)).
            Count(ctx)
        
        if err!= nil {
            log.Fatalf("Failed to query roles: %v", err)
        }

        if count == 0 {
            _, err := client.Role.Create().
                SetName(roleName).
                Save(ctx)
            if err!= nil {
                log.Fatalf("Failed to seed role %s: %v", roleName, err)
            }
            fmt.Printf("Role '%s' seeded successfully.\n", roleName)
        }
    }
}