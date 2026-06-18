package main

import (
	"fmt"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	err := ServerMux()
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
