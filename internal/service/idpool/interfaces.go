package idpool

type IDPoolRepo interface {
	GetIDPool(takeAmount uint64) (idPool *IDPool, err error)
	CreateNewIDPoolClaim(idPool *IDPool, takeAmount uint64) (res *IDPoolClaim, err error)
}
