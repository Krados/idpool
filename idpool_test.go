package idpool

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
)

// contendedOnceProvider returns false on the very first TryLock to exercise the
// sleep-and-retry branch, then behaves like a normal provider.
type contendedOnceProvider struct {
	mu    sync.Mutex
	tried bool
	ch    chan struct{}
}

func newContendedOnceProvider() *contendedOnceProvider {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return &contendedOnceProvider{ch: ch}
}

func (c *contendedOnceProvider) TryLock(_ string) (bool, error) {
	c.mu.Lock()
	firstCall := !c.tried
	c.tried = true
	c.mu.Unlock()
	if firstCall {
		return false, nil
	}
	select {
	case <-c.ch:
		return true, nil
	default:
		return false, nil
	}
}

func (c *contendedOnceProvider) Release(_ string) error {
	c.ch <- struct{}{}
	return nil
}

func (c *contendedOnceProvider) GetSet(_ string) (string, error) {
	return "1;1000", nil
}

// --- mock providers ---

// fixedRangeProvider is a lockable provider that always returns the same GetSet response.
type fixedRangeProvider struct {
	ch       chan struct{}
	response string
	err      error
}

func newFixedRangeProvider(response string, err error) *fixedRangeProvider {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return &fixedRangeProvider{ch: ch, response: response, err: err}
}

func (f *fixedRangeProvider) TryLock(_ string) (bool, error) {
	select {
	case <-f.ch:
		return true, nil
	default:
		return false, nil
	}
}

func (f *fixedRangeProvider) Release(_ string) error {
	f.ch <- struct{}{}
	return nil
}

func (f *fixedRangeProvider) GetSet(_ string) (string, error) {
	return f.response, f.err
}

// errTryLockProvider always returns an error from TryLock.
type errTryLockProvider struct{}

func (e *errTryLockProvider) TryLock(_ string) (bool, error) {
	return false, errors.New("trylock error")
}
func (e *errTryLockProvider) Release(_ string) error         { return nil }
func (e *errTryLockProvider) GetSet(_ string) (string, error) { return "", nil }

// --- IDPool tests ---

func TestNew_ReturnsNonNilPool(t *testing.T) {
	pool := New("key", NewLocalProvider())
	if pool == nil {
		t.Fatal("New returned nil")
	}
}

func TestIDPool_Get_ReturnsSequentialIDs(t *testing.T) {
	pool := New("key", NewLocalProvider())
	for want := 1; want <= 10; want++ {
		id, err := pool.Get()
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", want, err)
		}
		got, err := strconv.Atoi(id)
		if err != nil {
			t.Fatalf("iteration %d: expected numeric id, got %q", want, id)
		}
		if got != want {
			t.Errorf("iteration %d: expected %d, got %d", want, want, got)
		}
	}
}

func TestIDPool_Get_IDsRemainSequentialAcrossRangeBoundary(t *testing.T) {
	pool := New("key", NewLocalProvider())
	// 1005 calls spans the boundary between the first range (1-1000) and the second.
	const total = 1005
	prev := 0
	for i := 0; i < total; i++ {
		id, err := pool.Get()
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
		n, err := strconv.Atoi(id)
		if err != nil {
			t.Fatalf("call %d: expected numeric id, got %q", i+1, id)
		}
		if n != prev+1 {
			t.Errorf("call %d: expected %d, got %d (sequence broken)", i+1, prev+1, n)
		}
		prev = n
	}
}

func TestIDPool_Get_ConcurrentCallsProduceUniqueIDs(t *testing.T) {
	pool := New("key", NewLocalProvider())
	const goroutines = 50
	const perGoroutine = 20
	total := goroutines * perGoroutine

	results := make(chan string, total)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				id, err := pool.Get()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				results <- id
			}
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[string]bool, total)
	for id := range results {
		if seen[id] {
			t.Errorf("duplicate id %q", id)
		}
		seen[id] = true
	}
	if got := len(seen); got != total {
		t.Errorf("expected %d unique ids, got %d", total, got)
	}
}

func TestIDPool_Get_TryLockError_ReturnsError(t *testing.T) {
	pool := New("key", &errTryLockProvider{})
	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error from TryLock failure, got nil")
	}
}

func TestIDPool_Get_GetSetError_ReturnsError(t *testing.T) {
	pool := New("key", newFixedRangeProvider("", fmt.Errorf("storage error")))
	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error from GetSet failure, got nil")
	}
}

func TestIDPool_Get_InvalidResponse_NoSemicolon(t *testing.T) {
	pool := New("key", newFixedRangeProvider("invalid", nil))
	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error for response without semicolon, got nil")
	}
}

func TestIDPool_Get_InvalidResponse_BadStart(t *testing.T) {
	pool := New("key", newFixedRangeProvider("abc;1000", nil))
	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error for non-numeric start in response, got nil")
	}
}

func TestIDPool_Get_InvalidResponse_BadEnd(t *testing.T) {
	pool := New("key", newFixedRangeProvider("1;xyz", nil))
	_, err := pool.Get()
	if err == nil {
		t.Fatal("expected error for non-numeric end in response, got nil")
	}
}

func TestIDPool_Get_LargeNumberIDs(t *testing.T) {
	const bigStart = "99999999999999999999"
	const bigEnd = "100000000000000000000"
	pool := New("key", newFixedRangeProvider(bigStart+";"+bigEnd, nil))
	id, err := pool.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != bigStart {
		t.Errorf("expected %s, got %s", bigStart, id)
	}
}

func TestIDPool_Get_RetriesWhenLockContended(t *testing.T) {
	pool := New("key", newContendedOnceProvider())
	id, err := pool.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "1" {
		t.Errorf("expected \"1\", got %q", id)
	}
}
