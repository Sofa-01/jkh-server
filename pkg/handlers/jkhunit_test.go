// pkg/handlers/jkhunit_test.go

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

func setupJkhUnitTest(t *testing.T) (*gin.Engine, *ent.Client) {
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

	jkhService := service.NewJkhUnitService(client)
	jkhHandler := NewJkhUnitHandler(jkhService)

	r := gin.New()
	r.POST("/api/v1/jkhunits", jkhHandler.CreateJkhUnit)
	r.GET("/api/v1/jkhunits", jkhHandler.ListJkhUnits)
	r.GET("/api/v1/jkhunits/:id", jkhHandler.GetJkhUnit)
	r.PUT("/api/v1/jkhunits/:id", jkhHandler.UpdateJkhUnit)
	r.DELETE("/api/v1/jkhunits/:id", jkhHandler.DeleteJkhUnit)

	return r, client
}

func TestJkhUnitHandler_CreateJkhUnit_Success(t *testing.T) {
	r, client := setupJkhUnitTest(t)

	// Создаём район
	ctx := context.Background()
	district := client.District.Create().SetName("Район").SaveX(ctx)

	reqBody := models.CreateJkhUnitRequest{Name: "ЖЭУ-1", DistrictID: district.ID}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jkhunits", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.JkhUnitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Name != "ЖЭУ-1" {
		t.Errorf("Expected name 'ЖЭУ-1', got %s", resp.Name)
	}
}

func TestJkhUnitHandler_CreateJkhUnit_InvalidJSON(t *testing.T) {
	r, _ := setupJkhUnitTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jkhunits", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestJkhUnitHandler_ListJkhUnits(t *testing.T) {
	r, client := setupJkhUnitTest(t)

	ctx := context.Background()
	district := client.District.Create().SetName("Район").SaveX(ctx)
	client.JkhUnit.Create().SetName("ЖЭУ-1").SetDistrictID(district.ID).SaveX(ctx)
	client.JkhUnit.Create().SetName("ЖЭУ-2").SetDistrictID(district.ID).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jkhunits", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp []models.JkhUnitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 units, got %d", len(resp))
	}
}

func TestJkhUnitHandler_GetJkhUnit_Success(t *testing.T) {
	r, client := setupJkhUnitTest(t)

	ctx := context.Background()
	district := client.District.Create().SetName("Район").SaveX(ctx)
	client.JkhUnit.Create().SetName("Тест").SetDistrictID(district.ID).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jkhunits/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestJkhUnitHandler_GetJkhUnit_NotFound(t *testing.T) {
	r, _ := setupJkhUnitTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jkhunits/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestJkhUnitHandler_DeleteJkhUnit_Success(t *testing.T) {
	r, client := setupJkhUnitTest(t)

	ctx := context.Background()
	district := client.District.Create().SetName("Район").SaveX(ctx)
	client.JkhUnit.Create().SetName("Удалить").SetDistrictID(district.ID).SaveX(ctx)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jkhunits/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestJkhUnitHandler_DeleteJkhUnit_NotFound(t *testing.T) {
	r, _ := setupJkhUnitTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jkhunits/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

