package idpool

import (
	"sync"
)

type NoIDAvailable struct{}

func (e NoIDAvailable) Error() string {
	return "NoIDAvailable"
}

type IDPoolServ struct {
	repo           IDPoolRepo
	idQueue        chan uint64
	takeAmount     uint64
	availableCount int
	sync.Mutex
}

func NewIDPoolServ(repo IDPoolRepo) *IDPoolServ {
	serv := &IDPoolServ{
		repo:           repo,
		idQueue:        make(chan uint64, 1000),
		takeAmount:     1000,
		availableCount: 0,
	}

	return serv
}

func (s *IDPoolServ) Take() (uint64, error) {
	select {
	case id := <-s.idQueue:
		s.Lock()
		defer s.Unlock()
		s.availableCount--

		return id, nil
	default:
		s.Lock()
		defer s.Unlock()
		if s.availableCount > 0 {
			id := <-s.idQueue
			s.availableCount--
			return id, nil
		}

		idPoolClaim, err := s.createIDPoolClaim()
		if err != nil {
			return 0, err
		}
		for i := idPoolClaim.StartPos; i < idPoolClaim.EndPos; i++ {
			s.idQueue <- i
			s.availableCount++
		}

		id := <-s.idQueue
		s.availableCount--

		return id, nil
	}
}

func (s *IDPoolServ) createIDPoolClaim() (res *IDPoolClaim, err error) {
	idPool, err := s.repo.GetIDPool(s.takeAmount)
	if err != nil {
		return
	}

	idPoolClaim, err := s.repo.CreateNewIDPoolClaim(idPool, s.takeAmount)
	if err != nil {
		return
	}
	res = idPoolClaim

	return
}
