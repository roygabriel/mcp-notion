package notion

import (
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedStateAllowsRequests(t *testing.T) {
	cb := NewCircuitBreaker()

	if cb.State() != CircuitClosed {
		t.Fatalf("new circuit breaker should start closed, got state %d", cb.State())
	}

	if !cb.Allow() {
		t.Error("closed circuit breaker should allow requests")
	}
}

func TestCircuitBreaker_OpensAfterThresholdFailures(t *testing.T) {
	cb := NewCircuitBreaker()

	// Record failures below threshold — should remain closed
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
		if cb.State() != CircuitClosed {
			t.Fatalf("circuit should remain closed after %d failures, got state %d", i+1, cb.State())
		}
	}

	// 5th failure hits threshold — should open
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("circuit should be open after 5 failures, got state %d", cb.State())
	}
}

func TestCircuitBreaker_OpenStateBlocksRequests(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker()
	cb.now = func() time.Time { return now }

	// Trip the breaker
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen, got %d", cb.State())
	}

	if cb.Allow() {
		t.Error("open circuit breaker should block requests")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker()
	cb.now = func() time.Time { return now }

	// Trip the breaker
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen, got %d", cb.State())
	}

	// Advance time past the reset timeout (30s)
	now = now.Add(31 * time.Second)

	if !cb.Allow() {
		t.Error("circuit breaker should allow a probe request after reset timeout")
	}

	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen after timeout, got %d", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker()
	cb.now = func() time.Time { return now }

	// Trip the breaker, then advance past timeout to reach half-open
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}
	now = now.Add(31 * time.Second)
	cb.Allow() // triggers transition to half-open

	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen, got %d", cb.State())
	}

	// Success in half-open should close the circuit
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after half-open success, got %d", cb.State())
	}

	if !cb.Allow() {
		t.Error("circuit should allow requests after returning to closed")
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker()
	cb.now = func() time.Time { return now }

	// Trip the breaker, then advance past timeout to reach half-open
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}
	now = now.Add(31 * time.Second)
	cb.Allow() // triggers transition to half-open

	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen, got %d", cb.State())
	}

	// Failure in half-open should re-open the circuit.
	// The failure count was already at threshold; one more pushes it past again.
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen after half-open failure, got %d", cb.State())
	}

	// Should block requests again (time hasn't advanced further)
	if cb.Allow() {
		t.Error("re-opened circuit breaker should block requests")
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := NewCircuitBreaker()

	// Accumulate 4 failures (one below threshold)
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
	}

	// Success should reset the count
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after success, got %d", cb.State())
	}

	// Another 4 failures should not open the circuit (count was reset)
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitClosed {
		t.Fatalf("circuit should remain closed after 4 failures following a reset, got %d", cb.State())
	}

	// But the 5th failure after the reset should open it
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Fatalf("circuit should be open after 5 consecutive failures, got %d", cb.State())
	}
}

func TestCircuitBreaker_TimeoutBoundary(t *testing.T) {
	tests := []struct {
		name      string
		elapsed   time.Duration
		wantAllow bool
		wantState CircuitState
	}{
		{
			name:      "before timeout",
			elapsed:   29 * time.Second,
			wantAllow: false,
			wantState: CircuitOpen,
		},
		{
			name:      "exactly at timeout",
			elapsed:   30 * time.Second,
			wantAllow: false,
			wantState: CircuitOpen,
		},
		{
			name:      "just past timeout",
			elapsed:   30*time.Second + time.Millisecond,
			wantAllow: true,
			wantState: CircuitHalfOpen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			cb := NewCircuitBreaker()
			cb.now = func() time.Time { return now }

			for i := 0; i < 5; i++ {
				cb.RecordFailure()
			}

			now = now.Add(tt.elapsed)

			got := cb.Allow()
			if got != tt.wantAllow {
				t.Errorf("Allow() = %v, want %v", got, tt.wantAllow)
			}

			if cb.State() != tt.wantState {
				t.Errorf("State() = %d, want %d", cb.State(), tt.wantState)
			}
		})
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker()
	var wg sync.WaitGroup

	// Spawn goroutines that concurrently call all methods.
	// Run under -race to detect data races.
	const goroutines = 50

	wg.Add(goroutines * 4)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cb.Allow()
		}()
		go func() {
			defer wg.Done()
			cb.RecordFailure()
		}()
		go func() {
			defer wg.Done()
			cb.RecordSuccess()
		}()
		go func() {
			defer wg.Done()
			cb.State()
		}()
	}

	wg.Wait()
}

func TestCircuitBreaker_NewCircuitBreakerDefaults(t *testing.T) {
	cb := NewCircuitBreaker()

	if cb.threshold != 5 {
		t.Errorf("threshold = %d, want 5", cb.threshold)
	}

	if cb.resetTimeout != 30*time.Second {
		t.Errorf("resetTimeout = %v, want 30s", cb.resetTimeout)
	}

	if cb.state != CircuitClosed {
		t.Errorf("initial state = %d, want CircuitClosed (%d)", cb.state, CircuitClosed)
	}

	if cb.now == nil {
		t.Error("now function should not be nil")
	}
}
