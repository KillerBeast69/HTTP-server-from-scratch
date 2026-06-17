package main

import (
	"fmt"
)

func main() {
	err := ServerMux()
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}