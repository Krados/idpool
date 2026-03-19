package idpool

import (
	"sync"
	"testing"
)

func TestNewLocalProvider(t *testing.T) {
	p := NewLocalProvider()
	if p == nil {
		t.Fatal("NewLocalProvider returned nil")
	}
}

func TestLocalProvider_TryLock_SucceedsOnFreshProvider(t *testing.T) {
	p := NewLocalProvider()
	ok, err := p.TryLock("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected TryLock to return true on a fresh provider")
	}
}

func TestLocalProvider_TryLock_FailsWhenAlreadyLocked(t *testing.T) {
	p := NewLocalProvider()
	if _, err := p.TryLock("key"); err != nil {
		t.Fatalf("first TryLock returned unexpected error: %v", err)
	}
	ok, err := p.TryLock("key")
	if err != nil {
		t.Fatalf("second TryLock returned unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected TryLock to return false while already locked")
	}
}

func TestLocalProvider_Release_ReenablesLock(t *testing.T) {
	p := NewLocalProvider()
	if _, err := p.TryLock("key"); err != nil {
		t.Fatalf("TryLock returned unexpected error: %v", err)
	}
	if err := p.Release("key"); err != nil {
		t.Fatalf("Release returned unexpected error: %v", err)
	}
	ok, err := p.TryLock("key")
	if err != nil {
		t.Fatalf("TryLock after Release returned unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected TryLock to succeed after Release")
	}
}

func TestLocalProvider_GetSet_FirstCall_Returns_1_To_1000(t *testing.T) {
	p := NewLocalProvider()
	got, err := p.GetSet("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "1;1000" {
		t.Errorf("expected \"1;1000\", got %q", got)
	}
}

func TestLocalProvider_GetSet_ReturnsConsecutiveRanges(t *testing.T) {
	p := NewLocalProvider()
	expected := []string{"1;1000", "1001;2000", "2001;3000"}
	for i, want := range expected {
		got, err := p.GetSet("key")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
		if got != want {
			t.Errorf("call %d: expected %q, got %q", i+1, want, got)
		}
	}
}

func TestLocalProvider_GetSet_DifferentKeysHaveIndependentRanges(t *testing.T) {
	p := NewLocalProvider()

	got, err := p.GetSet("a")
	if err != nil {
		t.Fatalf("key a, call 1: %v", err)
	}
	if got != "1;1000" {
		t.Errorf("key a, call 1: expected \"1;1000\", got %q", got)
	}

	got, err = p.GetSet("b")
	if err != nil {
		t.Fatalf("key b, call 1: %v", err)
	}
	if got != "1;1000" {
		t.Errorf("key b, call 1: expected \"1;1000\", got %q", got)
	}

	got, err = p.GetSet("a")
	if err != nil {
		t.Fatalf("key a, call 2: %v", err)
	}
	if got != "1001;2000" {
		t.Errorf("key a, call 2: expected \"1001;2000\", got %q", got)
	}
}

func TestLocalProvider_GetSet_ConcurrentSafety(t *testing.T) {
	p := NewLocalProvider()
	const n = 20

	// Each goroutine writes to its own index – no data race on the slice.
	results := make([]string, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			val, err := p.GetSet("key")
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", idx, err)
				return
			}
			results[idx] = val
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for _, r := range results {
		if seen[r] {
			t.Errorf("duplicate range %q returned by concurrent GetSet", r)
		}
		seen[r] = true
	}
	if len(seen) != n {
		t.Errorf("expected %d unique ranges, got %d", n, len(seen))
	}
}

func TestLocalProvider_GetSet_InvalidStoredValue_ReturnsError(t *testing.T) {
	p := NewLocalProvider().(*LocalProvider)
	// Seed the map with a non-decimal value to trigger the error branch.
	p.m["key"] = "not-a-number"

	_, err := p.GetSet("key")
	if err == nil {
		t.Fatal("expected error for invalid stored value, got nil")
	}
}
