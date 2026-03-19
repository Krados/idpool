package idpool

type Provider interface {
	// GetSet should be concurrent safe, and return a string in the format of "start;end",
	// where start and end are decimal numbers.
	GetSet(key string) (string, error)
	// TryLock should be concurrent safe, and return true if the lock is acquired successfully.
	TryLock(key string) (bool, error)
	// Release should be concurrent safe, and release the lock acquired by TryLock.
	Release(key string) error
}
