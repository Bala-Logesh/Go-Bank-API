package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world")

	server := NewAPIServer(":3000")
	server.Run()
}
