// Written by Gemini CLI
// This file is licensed under the MIT License.
// See the LICENSE file for details.

//go:build chaos
// +build chaos

package chaos_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock" // Import go-sqlmock
	"github.com/stevemcghee/go-to-production/internal/app"
	_ "github.com/lib/pq"
	"github.com/sony/gobreaker"
	"github.com/cenkalti/backoff/v4"
)

var (
	mocksql sqlmock.Sqlmock
	mockdb  *sql.DB
)

// TestMain sets up and tears down the test database using go-sqlmock.
func TestMain(m *testing.M) {
	var err error
	// Use sqlmock.MonitorPingsOption(true) to allow ExpectPing
	mockdb, mocksql, err = sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		fmt.Printf("Failed to create sqlmock: %v\n", err)
		os.Exit(1)
	}
	defer mockdb.Close() // Close mockdb at the end of TestMain

	// Save original app.DB, app.DBRead, and app.backoffStrategy state
	originalAppDB := app.DB
	originalAppDBRead := app.DBRead
	originalBackoffStrategy := app.BackoffStrategy

	// Use a fixed backoff for testing: 1 initial attempt + 2 retries = 3 attempts total
	app.BackoffStrategy = backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Millisecond), 2) 

	defer func() {
		// Restore original app.DB, app.DBRead, and app.backoffStrategy state
		app.DB = originalAppDB
		app.DBRead = originalAppDBRead
		app.BackoffStrategy = originalBackoffStrategy
	}()

	// Set app.DB variables for handlers to use mockdb
	// These will be restored after all tests in this package run
	app.DB = mockdb
	app.DBRead = mockdb // Initially point both to mockdb

	// Run tests
	code := m.Run()

	// Ensure all expectations were met at the end of the entire test suite
	// This captures any lingering expectations not cleared by individual tests
	if err := mocksql.ExpectationsWereMet(); err != nil {
		fmt.Printf("there were unfulfilled expectations at the end of TestMain: %s\n", err)
	}
	// mockdb.Close() is already deferred at the beginning of the function
	os.Exit(code)
}

// TestChaosHealthzDBConnectionFailure tests /healthz when DB connection pings fail.
func TestChaosHealthzDBConnectionFailure(t *testing.T) {
	t.Cleanup(func() { // Ensures expectations are met even if test fails mid-way
		if err := mocksql.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations at the end of TestChaosHealthzDBConnectionFailure: %s", err)
		}
	})

	// Simulate DB connection failure by making mockdb return an error on Ping
	mocksql.ExpectPing().WillReturnError(fmt.Errorf("simulated db connection error"))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	app.HealthzHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on /healthz with db down, got %d", http.StatusInternalServerError, w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Database connection failed")) {
		t.Errorf("expected body to contain 'Database connection failed', got %q", w.Body.String())
	}
}

// TestChaosGetTodosDBConnectionFailure tests /todos when DB queries fail due to connection issues.
func TestChaosGetTodosDBConnectionFailure(t *testing.T) {
	t.Cleanup(func() { // Ensures expectations are met even if test fails mid-way
		if err := mocksql.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations at the end of TestChaosGetTodosDBConnectionFailure: %s", err)
		}
	})

	// Save original app.CB state
	originalAppCB := app.CB
	defer func() {
		app.CB = originalAppCB
	}()

	// Create a temporary CB that uses a very short retry timeout to speed up the test
	tempCB := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "TempGetTodosCB",
		Timeout: 50 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return true // Trip immediately on first failure
		},
	})
	app.CB = tempCB

	// Simulate query failure multiple times (due to retry mechanism and fallback)
	// app.RetryOperation retries up to MaxElapsedTime (5s). With InitialInterval=100ms, MaxInterval=2s.
	// This means a maximum of about 7 attempts (initial + 6 retries) if each takes 100ms+. Plus fallback.
	// Set enough expectations to cover both app.DBRead and app.DB attempts during a full retry cycle.
	numExpectedFailures := 3
	for i := 0; i < numExpectedFailures; i++ {
		mocksql.ExpectQuery("SELECT (.+) FROM todos").WillReturnError(fmt.Errorf("simulated db query error"))
	}

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	w := httptest.NewRecorder()
	app.GetTodos(w, req)

	// Due to retry mechanism, it should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on GetTodos with db down, got %d", http.StatusInternalServerError, w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("simulated db query error")) {
		t.Errorf("expected body to contain 'simulated db query error', got %q", w.Body.String())
	}
}


// TestChaosCircuitBreakerOpensAndCloses simulates a scenario where the database becomes unavailable
// and then available again, testing if the circuit breaker opens and then closes.
func TestChaosCircuitBreakerOpensAndCloses(t *testing.T) {
	t.Cleanup(func() { // Ensures expectations are met even if test fails mid-way
		if err := mocksql.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations at the end of TestChaosCircuitBreakerOpensAndCloses: %s", err)
		}
	})

	// Use a test-specific circuit breaker
	var st gobreaker.Settings
	st.Name = "TestChaosCB"
	st.MaxRequests = 1
	st.Interval = 1 * time.Second // Short interval to speed up test
	st.Timeout = 2 * time.Second  // Short timeout to speed up test
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// Trip after 2 consecutive failures
		return counts.ConsecutiveFailures >= 2
	}
	testCB := gobreaker.NewCircuitBreaker(st)

	// Save original app.CB state
	originalAppCB := app.CB
	defer func() {
		app.CB = originalAppCB
	}()
	app.CB = testCB

	// --- Phase 1: DB is down, trip the circuit breaker ---
	numExpectedFailuresPerLogicalCall := 3 // (initial + 2 retries)
	
	// Add logging to circuit breaker state changes and ReadyToTrip for debugging
	st.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		t.Logf("Circuit Breaker state changed: %s from %s to %s", name, from, to)
	}
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		t.Logf("ReadyToTrip: Requests=%d, TotalFailures=%d, ConsecutiveFailures=%d", counts.Requests, counts.TotalFailures, counts.ConsecutiveFailures)
		return counts.ConsecutiveFailures >= 2
	}
	
	// We need 2 logical failures to trip the CB (ReadyToTrip = ConsecutiveFailures >= 2).
	// So, we need to make 2 logical calls to app.GetTodos, each failing after retries.
	for i := 0; i < 2 * numExpectedFailuresPerLogicalCall; i++ {
		mocksql.ExpectQuery("SELECT (.+) FROM todos").WillReturnError(fmt.Errorf("simulated db query error CB"))
	}

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	w := httptest.NewRecorder()
	app.GetTodos(w, req) // First logical failure from CB perspective

		w = httptest.NewRecorder()

		app.GetTodos(w, req) // Second logical failure from CB perspective

	

		// Check state immediately after the second failure that should trip it

		if testCB.State() != gobreaker.StateOpen {

			t.Fatalf("circuit breaker should be open after consecutive failures, state is %s", testCB.State().String())

		}

	// --- Phase 2: DB is still down, confirm requests are blocked ---
	// Expect no query from mock as circuit is open. The CB will return ErrOpenState
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	w = httptest.NewRecorder()
	app.GetTodos(w, req) // This call should be blocked by CB

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 when circuit is open, got %d", w.Code)
	}

	// --- Phase 3: Wait for CB to enter half-open state ---
	t.Logf("Waiting for %v for circuit breaker to enter half-open state...", st.Timeout)
	time.Sleep(st.Timeout + 500*time.Millisecond) // Add a small buffer

	if testCB.State() != gobreaker.StateHalfOpen {
		t.Fatalf("circuit breaker should be half-open after timeout, state is %s", testCB.State().String())
	}

	// --- Phase 4: DB comes back up, test recovery ---
	t.Log("Restoring database connection (mocksql to return success)...")
	// Configure mocksql to return a successful query for the single request in half-open state
	mocksql.ExpectQuery("SELECT (.+) FROM todos").WillReturnRows(sqlmock.NewRows([]string{"id", "task", "completed"}).AddRow(1, "Test Task", false))

	// This request in half-open state should succeed and close the circuit
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	w = httptest.NewRecorder()
	app.GetTodos(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d on recovery request, got %d. Body: %q", http.StatusOK, w.Code, w.Body.String())
	}

	// --- Phase 5: Confirm CB is closed again ---
	if testCB.State() != gobreaker.StateClosed {
		t.Fatalf("circuit breaker should be closed after a successful request, state is %s", testCB.State().String())
	}

	// Subsequent requests should also succeed
	mocksql.ExpectQuery("SELECT (.+) FROM todos").WillReturnRows(sqlmock.NewRows([]string{"id", "task", "completed"}).AddRow(2, "Another Task", true))
	req = httptest.NewRequest(http.MethodGet, "/todos", nil)
	w = httptest.NewRecorder()
	app.GetTodos(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d on subsequent request, got %d", http.StatusOK, w.Code) // Changed %q to %d
	}
}

func TestChaosReadReplicaFailover(t *testing.T) {
	// Create separate mock DBs for primary and read replica
	mockdbPrimary, mocksqlPrimary, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create primary sqlmock: %v", err)
	}
	defer mockdbPrimary.Close()

	mockdbReplica, mocksqlReplica, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create replica sqlmock: %v", err)
	}
	defer mockdbReplica.Close()

	t.Cleanup(func() {
		// Ensure expectations are met for both mocks
		if err := mocksqlPrimary.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations for primary mock at the end of TestChaosReadReplicaFailover: %s", err)
		}
		if err := mocksqlReplica.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations for replica mock at the end of TestChaosReadReplicaFailover: %s", err)
		}
	})

	// Save original app.DB and app.DBRead state
	originalAppDB := app.DB
	originalAppDBRead := app.DBRead
	defer func() {
		app.DB = originalAppDB
		app.DBRead = originalAppDBRead
	}()

	// Configure app.DB to use mockdbPrimary (primary success)
	app.DB = mockdbPrimary
	// Configure app.DBRead to use mockdbReplica (replica failure)
	app.DBRead = mockdbReplica

	// Expect the query to mockdbReplica to fail multiple times due to retries
	// `RetryOperation` attempts 8 times
	numReadReplicaFailures := 1
	for i := 0; i < numReadReplicaFailures; i++ {
		mocksqlReplica.ExpectQuery("SELECT id, task, completed FROM todos ORDER BY id").WillReturnError(fmt.Errorf("simulated read replica failure"))
	}

	// Expect the subsequent query to mockdbPrimary to succeed (after replica failures and fallback)
	mocksqlPrimary.ExpectQuery("SELECT id, task, completed FROM todos ORDER BY id").WillReturnRows(sqlmock.NewRows([]string{"id", "task", "completed"}).AddRow(2, "Fallback Task", true))


	// Make a GET request, which should use the read replica first, fail, and fall back to the primary
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	w := httptest.NewRecorder()
	app.GetTodos(w, req)

	// The request should succeed by falling back to the primary
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d when read replica is down but primary is up, got %d. Body: %q", http.StatusOK, w.Code, w.Body.String())
	}
}