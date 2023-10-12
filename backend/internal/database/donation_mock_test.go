//go:build !integration
// +build !integration

package database

import (
	"testing"
)

func TestCreateDonation(t *testing.T) {
	mockRepo := NewDonationMock()

	donation := Donation{
		StreamerID: 1,
		Status:     DonationStatusCreated,
		PaymentID:  "12345",
	}

	err := mockRepo.Create(&donation)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if donation.ID != 1 {
		t.Errorf("Create didn't set the ID correctly, expected 1, got %d", donation.ID)
	}
}

func TestGetDonation(t *testing.T) {
	mockRepo := NewDonationMock()

	donation := Donation{
		StreamerID: 1,
		Status:     DonationStatusCreated,
		PaymentID:  "12345",
	}

	mockRepo.Create(&donation)

	retrievedDonation, err := mockRepo.GetDonation(1)
	if err != nil {
		t.Errorf("GetDonation failed: %v", err)
	}

	if retrievedDonation.ID != 1 {
		t.Errorf("GetDonation returned an incorrect donation")
	}

	_, err = mockRepo.GetDonation(2)
	if err == nil || err.Error() != "Donation not found" {
		t.Errorf("GetDonation should return an error for a non-existent donation")
	}
}

func TestUpdateDonation(t *testing.T) {
	mockRepo := NewDonationMock()

	donation := Donation{
		StreamerID: 1,
		Status:     DonationStatusCreated,
		PaymentID:  "12345",
	}

	mockRepo.Create(&donation)

	updatedDonation := Donation{
		ID:         1,
		StreamerID: 2,
		Status:     DonationStatusPayed,
		PaymentID:  "54321",
	}

	err := mockRepo.Update(updatedDonation)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	updatedDonation, _ = mockRepo.GetDonation(1)
	if updatedDonation.StreamerID != 2 || updatedDonation.Status != DonationStatusPayed || updatedDonation.PaymentID != "54321" {
		t.Errorf("Update didn't update the donation correctly")
	}

	err = mockRepo.Update(Donation{ID: 2})
	if err == nil || err.Error() != "Donation not found" {
		t.Errorf("Update should return an error for a non-existent donation")
	}
}

func TestGetDonations(t *testing.T) {
	mockRepo := NewDonationMock()

	donation1 := Donation{
		StreamerID: 1,
		Status:     DonationStatusCreated,
		PaymentID:  "12345",
	}

	donation2 := Donation{
		StreamerID: 2,
		Status:     DonationStatusPayed,
		PaymentID:  "54321",
	}

	mockRepo.Create(&donation1)
	mockRepo.Create(&donation2)

	// Test getting all donations
	donations, err := mockRepo.GetDonations(nil)
	if err != nil {
		t.Errorf("GetDonations failed: %v", err)
	}

	if len(donations) != 2 {
		t.Errorf("GetDonations should return all donations, expected 2, got %d", len(donations))
	}

	// Test getting donations by specific criteria
	filter := Donation{StreamerID: 1, Status: DonationStatusCreated}
	filteredDonations, err := mockRepo.GetDonations(&filter)
	if err != nil {
		t.Errorf("GetDonations with filter failed: %v", err)
	}

	if len(filteredDonations) != 1 {
		t.Errorf("GetDonations with filter should return 1 donations, got %d", len(filteredDonations))
	}
}
