// pkg/db/client.go

package db

import (
	"context"
	"fmt"
	"log"

	// Импортируем сгенерированный клиент Ent
	"jkh/ent" 

	// Драйвер для PostgreSQL [1]
	_ "github.com/lib/pq" 

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
)

//!!! ВАЖНО: ЗАМЕНИТЕ ЭТИ ЗНАЧЕНИЯ НА ДАННЫЕ ВАШЕГО POSTGRESQL!!!
const dsn = "host=localhost port=5432 user=postgres password=23052006 dbname=jkh-inspection sslmode=disable"

// NewClient создает и инициализирует клиент Ent
func NewClient() *ent.Client {
	client, err := ent.Open(dialect.Postgres, dsn)
	if err!= nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	// Миграции будут выполняться при первом запуске
	ctx := context.Background()
	
	// Метод ApplyAll создает все таблицы, ФК и индексы, используя нашу схему (10 таблиц, 3НФ)
	if err := client.Schema.Create(
		ctx, 
		schema.WithDropColumn(false), 
		schema.WithDropIndex(false),
	); err!= nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	fmt.Println("Database client and schema initialized successfully.")
	return client
}