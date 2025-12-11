// pkg/testutil/testutil.go

package testutil

import (
	"context"
	"database/sql"
	"testing"

	"jkh/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

// SetupTestDB создаёт тестовую БД SQLite in-memory
// Автоматически закрывается после завершения теста
func SetupTestDB(t *testing.T) *ent.Client {
	t.Helper()

	// Открываем SQLite in-memory с включённым foreign_keys
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Создаём ent-драйвер
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	// Создаём схему
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Создаём базовые роли для тестов
	roles := []string{"Specialist", "Coordinator", "Inspector"}
	for _, roleName := range roles {
		client.Role.Create().SetName(roleName).SaveX(ctx)
	}

	// Регистрируем cleanup
	t.Cleanup(func() {
		client.Close()
		db.Close()
	})

	return client
}

// SetupTestDBWithoutRoles создаёт тестовую БД без предустановленных ролей
func SetupTestDBWithoutRoles(t *testing.T) *ent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
		db.Close()
	})

	return client
}
