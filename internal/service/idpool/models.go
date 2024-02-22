package idpool

type IDPool struct {
	ID         uint64
	StartPos   uint64
	EndPos     uint64
	CurrentPos uint64
}

type IDPoolClaim struct {
	ID        uint64
	StartPos  uint64
	EndPos    uint64
	ClaimedAt uint64
}
