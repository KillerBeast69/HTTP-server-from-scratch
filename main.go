package main

import (
	"chirpy/internal/database"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	secret         string
	PolkaKey       string
}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("JWT_SECRET")
	polka_key := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	err = ServerMux(dbQueries, platform, secret, polka_key)
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
