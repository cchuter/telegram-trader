package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cchuter/telegram-trader/internal/blockchain"
	"github.com/cchuter/telegram-trader/internal/galachain"
	"github.com/cchuter/telegram-trader/internal/storage"
)

// HealthStatus represents the overall health of the service
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// DependencyStatus represents the status of a single dependency
type DependencyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       HealthStatus       `json:"status"`
	Timestamp    string             `json:"timestamp"`
	Dependencies []DependencyStatus `json:"dependencies"`
	Version      string             `json:"version"`
	Uptime       string             `json:"uptime"`
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	db         storage.Database
	tonClient  blockchain.Client
	galaClient *galachain.Client
	startTime  time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db storage.Database, tonClient blockchain.Client, galaClient *galachain.Client) *HealthChecker {
	return &HealthChecker{
		db:         db,
		tonClient:  tonClient,
		galaClient: galaClient,
		startTime:  time.Now(),
	}
}

// Check performs health checks on all dependencies
func (h *HealthChecker) Check(ctx context.Context) *HealthResponse {
	dependencies := []DependencyStatus{}

	// Check database
	dbStatus := h.checkDatabase(ctx)
	dependencies = append(dependencies, dbStatus)

	// Check TON client
	tonStatus := h.checkTonClient(ctx)
	dependencies = append(dependencies, tonStatus)

	// Check GalaChain gRPC connection
	galaStatus := h.checkGalaChain(ctx)
	dependencies = append(dependencies, galaStatus)

	// Determine overall status
	overallStatus := h.calculateOverallStatus(dependencies)

	// Calculate uptime
	uptime := time.Since(h.startTime).Round(time.Second).String()

	return &HealthResponse{
		Status:       overallStatus,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: dependencies,
		Version:      "1.0.0",
		Uptime:       uptime,
	}
}

// checkDatabase checks database connectivity
func (h *HealthChecker) checkDatabase(ctx context.Context) DependencyStatus {
	start := time.Now()

	// Try to query a user session (this will test database connectivity)
	// Using userID 0 which likely doesn't exist, but will test the query
	_, err := h.db.GetUserSession(ctx, 0)

	latency := time.Since(start)

	// "not found" errors are expected and indicate the database is working
	if err != nil {
		errMsg := err.Error()
		if errMsg != "sql: no rows in result set" &&
			errMsg != "user session not found" &&
			errMsg != "user session not found for user_id 0" {
			return DependencyStatus{
				Name:    "database",
				Status:  "unhealthy",
				Message: fmt.Sprintf("Database connection failed: %v", err),
				Latency: latency.String(),
			}
		}
	}

	return DependencyStatus{
		Name:    "database",
		Status:  "healthy",
		Message: "Connected",
		Latency: latency.String(),
	}
}

// checkTonClient checks TON blockchain connectivity
func (h *HealthChecker) checkTonClient(ctx context.Context) DependencyStatus {
	start := time.Now()

	// Try to get balance for a well-known address (TON Foundation)
	// This tests the connection to the TON network
	testAddr := "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c" // TON native address

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.tonClient.GetBalance(ctx, testAddr)

	latency := time.Since(start)

	if err != nil {
		return DependencyStatus{
			Name:    "ton_client",
			Status:  "unhealthy",
			Message: fmt.Sprintf("TON client failed: %v", err),
			Latency: latency.String(),
		}
	}

	return DependencyStatus{
		Name:    "ton_client",
		Status:  "healthy",
		Message: "Connected to TON network",
		Latency: latency.String(),
	}
}

// checkGalaChain checks GalaChain gRPC connectivity
func (h *HealthChecker) checkGalaChain(ctx context.Context) DependencyStatus {
	if h.galaClient == nil {
		return DependencyStatus{
			Name:    "galachain_grpc",
			Status:  "unhealthy",
			Message: "GalaChain client not initialized",
		}
	}

	start := time.Now()

	// Try to call GetBalance (even with invalid user ID, this tests connectivity)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.galaClient.GetBalance(ctx, 0)

	latency := time.Since(start)

	if err != nil {
		return DependencyStatus{
			Name:    "galachain_grpc",
			Status:  "unhealthy",
			Message: fmt.Sprintf("gRPC connection failed: %v", err),
			Latency: latency.String(),
		}
	}

	return DependencyStatus{
		Name:    "galachain_grpc",
		Status:  "healthy",
		Message: "Connected to GalaChain service",
		Latency: latency.String(),
	}
}

// calculateOverallStatus determines the overall health based on dependencies
func (h *HealthChecker) calculateOverallStatus(dependencies []DependencyStatus) HealthStatus {
	unhealthyCount := 0
	degradedCount := 0

	for _, dep := range dependencies {
		if dep.Status == "unhealthy" {
			unhealthyCount++
		} else if dep.Status == "degraded" {
			degradedCount++
		}
	}

	// If multiple dependencies are unhealthy, overall is unhealthy
	if unhealthyCount >= 2 {
		return StatusUnhealthy
	}

	// If one dependency is unhealthy or any are degraded, overall is degraded
	if unhealthyCount > 0 || degradedCount > 0 {
		return StatusDegraded
	}

	// All dependencies are healthy
	return StatusHealthy
}

// HTTPHandler returns an HTTP handler for the health check endpoint
func (h *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Perform health check
		health := h.Check(ctx)

		// Set status code based on health
		statusCode := http.StatusOK
		if health.Status == StatusDegraded {
			statusCode = http.StatusOK // Still 200, but degraded
		} else if health.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		}
	}
}

// StartHealthServer starts an HTTP server for health checks
func (h *HealthChecker) StartHealthServer(port string) error {
	http.HandleFunc("/health", h.HTTPHandler())
	return http.ListenAndServe(":"+port, nil)
}
