# idpool

`idpool` is an extensible incremental ID generator for Go with pluggable backend providers.
It abstracts range allocation and locking behind the `Provider` interface, making it suitable for both single-node and distributed environments.

## Features

- Uses a local buffered channel to reduce frequent backend range allocations
- Supports very large numeric IDs via `math/big.Int`
- Pluggable `Provider` design for Redis, databases, KV stores, or custom backends
- Includes a simple in-memory `LocalProvider`

## Installation

```bash
go get github.com/Krados/idpool
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/Krados/idpool"
)

func main() {
	pool := idpool.New("mypool", idpool.NewLocalProvider())

	id, err := pool.Get()
	if err != nil {
		panic(err)
	}

	fmt.Println("new id:", id)
}
```

For a complete sample, see `example/main.go`.

## API Overview

### `idpool.New(key string, provider Provider) *IDPool`

Creates an ID pool instance.

- `key`: Pool identifier. The same key shares the same allocated ranges.
- `provider`: Your backend implementation.

### `(*IDPool).Get() (string, error)`

Returns a new ID as a string.

High-level behavior:

1. If the local buffer has IDs, return one immediately.
2. Otherwise, try to acquire a lock via `TryLock`.
3. After locking, call `GetSet` to fetch a range in `start;end` format.
4. Return the first ID and enqueue the rest into the local buffer.

## `Provider` Interface

You can provide your own backend by implementing the following interface:

```go
type Provider interface {
	// GetSet must be concurrency-safe and return "start;end".
	GetSet(key string) (string, error)

	// TryLock must be concurrency-safe and return true when lock is acquired.
	TryLock(key string) (bool, error)

	// Release must be concurrency-safe and release the lock from TryLock.
	Release(key string) error
}
```

### Provider Implementation Notes

- Make `GetSet` atomic to avoid overlapping ranges across nodes.
- Return decimal integer strings only, in `start;end` format.
- Ensure `TryLock` and `Release` are reliable and properly paired.

## Built-in `LocalProvider` Behavior

`LocalProvider` is an in-memory implementation intended for local testing.

- On first allocation for a key, returns `1;1000`.
- Each subsequent allocation advances by one range (1000 IDs per range).
- Not shared across processes, so it is not suitable for multi-node production use.

## Concurrency and Performance

- `IDPool` uses a buffered channel of size 1000 to cache IDs.
- If lock acquisition fails, it retries after sleeping for 50ms.
- `Get()` can be called by multiple goroutines (assuming provider safety).

## Notes

- The ID type returned by this package is `string`.
- For globally unique and highly available IDs, implement a robust distributed `Provider`.
