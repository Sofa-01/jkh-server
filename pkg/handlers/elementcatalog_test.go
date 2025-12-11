// pkg/handlers/elementcatalog_test.go

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jkh/ent"
	"jkh/pkg/models"
	"jkh/pkg/service"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func setupElementCatalogTest(t *testing.T) (*gin.Engine, *ent.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

	elemService := service.NewElementCatalogService(client)
	elemHandler := NewElementCatalogHandler(elemService)

	r := gin.New()
	r.POST("/api/v1/elements", elemHandler.CreateElement)
	r.GET("/api/v1/elements", elemHandler.ListElements)
	r.GET("/api/v1/elements/:id", elemHandler.GetElement)
	r.PUT("/api/v1/elements/:id", elemHandler.UpdateElement)
	r.DELETE("/api/v1/elements/:id", elemHandler.DeleteElement)

	return r, client
}

func TestElementCatalogHandler_CreateElement_Success(t *testing.T) {
	r, _ := setupElementCatalogTest(t)

	category := "Несущие конструкции"
	reqBody := models.CreateElementCatalogRequest{Name: "Фундамент", Category: &category}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/elements", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.ElementCatalogResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Name != "Фундамент" {
		t.Errorf("Expected name 'Фундамент', got %s", resp.Name)
	}
}

func TestElementCatalogHandler_CreateElement_InvalidJSON(t *testing.T) {
	r, _ := setupElementCatalogTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/elements", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestElementCatalogHandler_ListElements(t *testing.T) {
	r, client := setupElementCatalogTest(t)

	ctx := context.Background()
	client.ElementCatalog.Create().SetName("Кровля").SaveX(ctx)
	client.ElementCatalog.Create().SetName("Стены").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/elements", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp []models.ElementCatalogResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(resp))
	}
}

func TestElementCatalogHandler_GetElement_Success(t *testing.T) {
	r, client := setupElementCatalogTest(t)

	ctx := context.Background()
	client.ElementCatalog.Create().SetName("Тест").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/elements/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestElementCatalogHandler_GetElement_NotFound(t *testing.T) {
	r, _ := setupElementCatalogTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/elements/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestElementCatalogHandler_UpdateElement_Success(t *testing.T) {
	r, client := setupElementCatalogTest(t)

	ctx := context.Background()
	client.ElementCatalog.Create().SetName("Старое").SaveX(ctx)

	category := "Новая категория"
	reqBody := models.CreateElementCatalogRequest{Name: "Новое", Category: &category}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/elements/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestElementCatalogHandler_DeleteElement_Success(t *testing.T) {
	r, client := setupElementCatalogTest(t)

	ctx := context.Background()
	client.ElementCatalog.Create().SetName("Удалить").SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/elements/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestElementCatalogHandler_DeleteElement_NotFound(t *testing.T) {
	r, _ := setupElementCatalogTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/elements/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

