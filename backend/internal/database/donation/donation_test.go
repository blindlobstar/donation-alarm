//go:build integration
// +build integration

package donation

import (
	"log"
	"os"
	"testing"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDonationRepoIntegration tests the DonationRepo interface methods.
func TestDonationRepoIntegration(t *testing.T) {
	// Open a connection to the test database (replace with your database connection logic).
	db, err := sqlx.Connect("postgres", os.Getenv("BACKEND__CONNECTION_STRING"))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()
	repo := Repo{Repo: database.Repo{DB: db}}

	db.Exec("DROP DATABASE testdb")
	repo.Migrate()

	_, err = db.Exec(`
		INSERT INTO streamers (twitch_id, twitch_name, secret_code) VALUES ('twitch_id_1', 'first_streamer', 'some_secret_code_1');
		INSERT INTO streamers (twitch_id, twitch_name, secret_code) VALUES ('twitch_id_2', 'second_streamer', 'some_secret_code_2');
		INSERT INTO streamers (twitch_id, twitch_name, secret_code) VALUES ('twitch_id_3', 'third_streamer', 'some_secret_code_3');
		INSERT INTO streamers (twitch_id, twitch_name, secret_code) VALUES ('twitch_id_4', 'fourth_streamer', 'some_secret_code_4');
	`)
	if err != nil {
		log.Fatalf("error seeding db: %v", err)
	}

	// Test create
	donation := &Donation{
		PaymentID:  "test_payment_id",
		StreamerID: 1,
		Amount:     100,
		Message:    "Test donation",
		Name:       "Test User",
		Status:     DonationStatusCreated,
	}

	err = repo.Create(donation)
	if err != nil {
		t.Fatal(err)
	}
	// Check that the donation has been created and has a valid ID.
	if donation.ID == 0 {
		t.Errorf("Expected a non-zero ID, but got 0")
	}

	testDonations := []Donation{
		{
			PaymentID:  "payment1",
			StreamerID: 1,
			Amount:     50,
			Message:    "Donation 1",
			Name:       "User A",
			Status:     DonationStatusCreated,
		},
		{
			PaymentID:  "payment2",
			StreamerID: 2,
			Amount:     75,
			Message:    "Donation 2",
			Name:       "User B",
			Status:     DonationStatusProcessing,
		},
	}

	for _, d := range testDonations {
		err := repo.Create(&d)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test GetDonations with various filters.
	cases := []struct {
		filter Donation
		count  int
	}{
		{Donation{Status: DonationStatusCreated}, 2},
		{Donation{Status: DonationStatusProcessing}, 1},
		{Donation{StreamerID: 1}, 2},
	}

	for _, tc := range cases {
		donations, err := repo.GetDonations(&tc.filter)
		if err != nil {
			t.Fatal(err)
		}
		if len(donations) != tc.count {
			t.Errorf("Expected %d donations, but got %d", tc.count, len(donations))
		}
	}

	// Test GetDonation by ID.
	retrievedDonation, err := repo.GetDonation(donation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retrievedDonation.ID != donation.ID {
		t.Errorf("Expected donation with ID %d, but got ID %d", donation.ID, retrievedDonation.ID)
	}

	// Update the status of the donation.
	donation.Status = DonationStatusProcessing
	err = repo.Update(*donation)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the updated donation and check its status.
	retrievedDonation, err = repo.GetDonation(donation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retrievedDonation.Status != DonationStatusProcessing {
		t.Errorf("Expected status %s, but got %s", DonationStatusProcessing, retrievedDonation.Status)
	}
}
