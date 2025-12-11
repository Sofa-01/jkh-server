// pkg/handlers/district_test.go

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

func setupDistrictTest(t *testing.T) (*gin.Engine, *ent.Client) {
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

	districtService := service.NewDistrictService(client)
	districtHandler := NewDistrictHandler(districtService)

	r := gin.New()
	r.POST("/api/v1/districts", districtHandler.CreateDistrict)
	r.GET("/api/v1/districts", districtHandler.ListDistricts)
	r.GET("/api/v1/districts/:id", districtHandler.GetDistrict)
	r.PUT("/api/v1/districts/:id", districtHandler.UpdateDistrict)
	r.DELETE("/api/v1/districts/:id", districtHandler.DeleteDistrict)

	return r, client
}

func TestDistrictHandler_CreateDistrict_Success(t *testing.T) {
	r, _ := setupDistrictTest(t)

	reqBody := models.CreateDistrictRequest{Name: "Центральный район"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/districts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.DistrictResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Name != "Центральный район" {
		t.Errorf("Expected name 'Центральный район', got %s", resp.Name)
	}
}

func TestDistrictHandler_CreateDistrict_InvalidJSON(t *testing.T) {
	r, _ := setupDistrictTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/districts", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDistrictHandler_ListDistricts(t *testing.T) {
	r, client := setupDistrictTest(t)

	// Создаём несколько районов напрямую в БД
	ctx := context.Background()
	client.District.Create().SetName("Район 1").SaveX(ctx)
	client.District.Create().SetName("Район 2").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/districts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp []models.DistrictResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 districts, got %d", len(resp))
	}
}

func TestDistrictHandler_GetDistrict_Success(t *testing.T) {
	r, client := setupDistrictTest(t)

	ctx := context.Background()
	d := client.District.Create().SetName("Тестовый").SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/districts/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.DistrictResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Name != d.Name {
		t.Errorf("Expected name %s, got %s", d.Name, resp.Name)
	}
}

func TestDistrictHandler_GetDistrict_NotFound(t *testing.T) {
	r, _ := setupDistrictTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/districts/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDistrictHandler_GetDistrict_InvalidID(t *testing.T) {
	r, _ := setupDistrictTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/districts/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDistrictHandler_UpdateDistrict_Success(t *testing.T) {
	r, client := setupDistrictTest(t)

	ctx := context.Background()
	client.District.Create().SetName("Старое имя").SaveX(ctx)

	reqBody := models.CreateDistrictRequest{Name: "Новое имя"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/districts/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.DistrictResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Name != "Новое имя" {
		t.Errorf("Expected name 'Новое имя', got %s", resp.Name)
	}
}

func TestDistrictHandler_DeleteDistrict_Success(t *testing.T) {
	r, client := setupDistrictTest(t)

	ctx := context.Background()
	client.District.Create().SetName("Для удаления").SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/districts/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestDistrictHandler_DeleteDistrict_NotFound(t *testing.T) {
	r, _ := setupDistrictTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/districts/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

