package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/storage"
)

func TestHealthCheck(t *testing.T) {
	// Initialize database
	db, err := storage.InitDB("./test_health.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	// Initialize TON client (connection will fail, but that's ok for testing)
	tonClient := ton.NewClient()

	// Initialize health checker
	healthChecker := NewHealthChecker(db, tonClient, nil)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Call health handler
	handler := healthChecker.HTTPHandler()
	handler(rec, req)

	// Check status code
	if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 200 or 503, got %d", rec.Code)
	}

	// Parse response
	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status == "" {
		t.Error("Status is empty")
	}

	if response.Timestamp == "" {
		t.Error("Timestamp is empty")
	}

	if len(response.Dependencies) == 0 {
		t.Error("No dependencies checked")
	}

	// Verify all expected dependencies are present
	expectedDeps := map[string]bool{
		"database":       false,
		"ton_client":     false,
		"galachain_grpc": false,
	}

	for _, dep := range response.Dependencies {
		if _, exists := expectedDeps[dep.Name]; exists {
			expectedDeps[dep.Name] = true
		}
	}

	for name, found := range expectedDeps {
		if !found {
			t.Errorf("Dependency %s not found in health check", name)
		}
	}

	t.Logf("Health check response: %+v", response)
}

func TestCalculateOverallStatus(t *testing.T) {
	healthChecker := &HealthChecker{
		startTime: time.Now(),
	}

	tests := []struct {
		name         string
		dependencies []DependencyStatus
		expected     HealthStatus
	}{
		{
			name: "all healthy",
			dependencies: []DependencyStatus{
				{Name: "db", Status: "healthy"},
				{Name: "ton", Status: "healthy"},
				{Name: "gala", Status: "healthy"},
			},
			expected: StatusHealthy,
		},
		{
			name: "one unhealthy - degraded",
			dependencies: []DependencyStatus{
				{Name: "db", Status: "healthy"},
				{Name: "ton", Status: "unhealthy"},
				{Name: "gala", Status: "healthy"},
			},
			expected: StatusDegraded,
		},
		{
			name: "two unhealthy - unhealthy",
			dependencies: []DependencyStatus{
				{Name: "db", Status: "unhealthy"},
				{Name: "ton", Status: "unhealthy"},
				{Name: "gala", Status: "healthy"},
			},
			expected: StatusUnhealthy,
		},
		{
			name: "one degraded - degraded",
			dependencies: []DependencyStatus{
				{Name: "db", Status: "healthy"},
				{Name: "ton", Status: "degraded"},
				{Name: "gala", Status: "healthy"},
			},
			expected: StatusDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := healthChecker.calculateOverallStatus(tt.dependencies)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHealthCheckDatabaseOnly(t *testing.T) {
	// Initialize database
	db, err := storage.InitDB("./test_health_db.db")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	healthChecker := &HealthChecker{
		db:        db,
		startTime: time.Now(),
	}

	ctx := context.Background()
	status := healthChecker.checkDatabase(ctx)

	if status.Name != "database" {
		t.Errorf("Expected name 'database', got '%s'", status.Name)
	}

	if status.Status != "healthy" {
		t.Errorf("Expected healthy database, got '%s': %s", status.Status, status.Message)
	}

	if status.Latency == "" {
		t.Error("Latency should be set")
	}

	t.Logf("Database check: %+v", status)
}
