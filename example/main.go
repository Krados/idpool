package main

import (
	"github.com/Krados/idpool"
)

func main() {
	// This example demonstrates how to use the IDPool with a LocalProvider.
	// you can implement your own provider by implementing the Provider interface
	// defined in provider.go.
	pool := idpool.New("mypool", idpool.NewLocalProvider())
	for i := 0; i < 10000001; i++ {
		pool.Get()
	}
}
