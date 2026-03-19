package idpool

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

type IDPool struct {
	key      string
	provider Provider
	ch       chan string
}

func New(key string, provider Provider) *IDPool {
	return &IDPool{
		key:      key,
		provider: provider,
		ch:       make(chan string, 1000),
	}
}

func (p *IDPool) Get() (string, error) {
	for {
		select {
		case id := <-p.ch:
			return id, nil
		default:
			ok, err := p.provider.TryLock(p.key)
			if err != nil {
				return "", err
			}
			if !ok {
				time.Sleep(time.Millisecond * 50)
				continue
			}
			defer p.provider.Release(p.key)
			raw, err := p.provider.GetSet(p.key)
			if err != nil {
				return "", err
			}
			parts := strings.Split(raw, ";")
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid provider response: %s", raw)
			}
			start := new(big.Int)
			end := new(big.Int)
			if _, ok := start.SetString(parts[0], 10); !ok {
				return "", fmt.Errorf("invalid start number: %s", parts[0])
			}
			if _, ok := end.SetString(parts[1], 10); !ok {
				return "", fmt.Errorf("invalid end number: %s", parts[1])
			}
			answer := start.String()
			start.Add(start, big.NewInt(1))
			for start.Cmp(end) <= 0 {
				p.ch <- start.String()
				start.Add(start, big.NewInt(1))
			}

			return answer, nil
		}
	}

}
