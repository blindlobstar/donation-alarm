//go:build integration
// +build integration

package database

import (
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Create a test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", os.Getenv("BACKEND__CONNECTION_STRING"))
	if err != nil {
		log.Printf("db: %s\n", os.Getenv("BACKEND__CONNECTION_STRING"))
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	// Create test table
	Repo{db}.Migrate()
	return db
}

func TestStreamerRepository(t *testing.T) {
	// Set up a test database connection
	db := setupTestDB(t)
	defer db.Close()
	db.Exec("DELETE * FROM donations")
	db.Exec("DELETE * FROM streamers")

	// Create a Repo with the test database connection
	repo := Repo{db}

	// Create a Streamer
	streamer := &Streamer{
		TwitchId:   "testTwitchId",
		TwitchName: "testTwitchName",
		SecretCode: "testSecretCode",
	}

	// Test CreateStreamer
	if err := repo.CreateStreamer(streamer); err != nil {
		t.Fatalf("Failed to create a streamer: %v", err)
	}

	// Test GetStreamers
	streamers, err := repo.GetStreamers(Streamer{TwitchName: "testTwitchName"})
	if err != nil {
		t.Fatalf("Failed to get streamers: %v", err)
	}

	if len(streamers) != 1 {
		t.Fatalf("Expected 1 streamer, got %d", len(streamers))
	}

	// Test GetStreamerById
	id := streamers[0].ID
	fetchedStreamer, err := repo.GetStreamerById(id)
	if err != nil {
		t.Fatalf("Failed to get streamer by ID: %v", err)
	}

	if fetchedStreamer.TwitchName != "testTwitchName" {
		t.Fatalf("Expected TwitchName to be 'testTwitchName', got %s", fetchedStreamer.TwitchName)
	}

	// Clean up test data
	_, err = db.Exec("DELETE FROM streamers WHERE id = $1", id)
	if err != nil {
		log.Printf("Failed to clean up test data: %v", err)
	}
}
