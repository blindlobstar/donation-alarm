package database

import "errors"

type DonationMock struct {
	donations map[int]Donation
	nextID    int
}

func NewDonationMock() *DonationMock {
	return &DonationMock{
		donations: make(map[int]Donation),
		nextID:    1,
	}
}

func (repo *DonationMock) Create(d *Donation) error {
	d.ID = repo.nextID
	repo.donations[d.ID] = *d
	repo.nextID++
	return nil
}

func (repo *DonationMock) GetDonations(d *Donation) ([]Donation, error) {
	var result []Donation
	for _, donation := range repo.donations {
		if d == nil || ((d.StreamerID == 0 || d.StreamerID == donation.StreamerID) &&
			(d.Status == "" || d.Status == donation.Status) &&
			(d.PaymentID == "" || d.PaymentID == donation.PaymentID)) {
			result = append(result, donation)
		}
	}
	return result, nil
}

func (repo *DonationMock) GetDonation(id int) (Donation, error) {
	donation, ok := repo.donations[id]
	if !ok {
		return Donation{}, errors.New("Donation not found")
	}
	return donation, nil
}

func (repo *DonationMock) Update(d Donation) error {
	_, ok := repo.donations[d.ID]
	if !ok {
		return errors.New("Donation not found")
	}
	repo.donations[d.ID] = d
	return nil
}
