package idpool

import (
	"fmt"
	"math/big"
	"sync"
)

type LocalProvider struct {
	m  map[string]string
	ch chan struct{}
	sync.Mutex
}

func (p *LocalProvider) TryLock(key string) (bool, error) {
	select {
	case <-p.ch:
		return true, nil
	default:
		return false, nil
	}
}

func (p *LocalProvider) Release(key string) error {
	p.ch <- struct{}{}
	return nil
}

func NewLocalProvider() Provider {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return &LocalProvider{
		m:  make(map[string]string),
		ch: ch,
	}
}

func (p *LocalProvider) GetSet(key string) (string, error) {
	p.Lock()
	defer p.Unlock()
	val, exist := p.m[key]

	if !exist {
		p.m[key] = "1000"
		return "1;1000", nil
	}
	n, ok := big.NewInt(0).SetString(val, 10)
	if !ok {
		return "", fmt.Errorf("invalid decimal number: %s", val)
	}

	answer := n.Add(n, big.NewInt(1)).String()
	maxV := n.Add(n, big.NewInt(999))
	answer += ";" + maxV.String()
	p.m[key] = maxV.String()

	return answer, nil
}
