package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

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
