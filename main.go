package main

import (
	"fmt"
	"log"
)

func main() {
	// Connecting to the database
	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Store connection: %+v\n", store)

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// Starting up the server
	server := NewAPIServer(":3000", store)
	server.Run()
}
