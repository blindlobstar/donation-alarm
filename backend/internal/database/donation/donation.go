package donation

import (
	"fmt"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
)

type Donation struct {
	PaymentID  string `db:"payment_id"`
	Message    string `db:"message"`
	Name       string `db:"name"`
	Status     string `db:"status"`
	ID         int
	StreamerID int `db:"streamer_id"`
	Amount     int `db:"amount"`
}

const (
	DonationStatusCreated    = "CREATED"
	DonationStatusProcessing = "PROCESSING"
	DonationStatusPayed      = "PAYED"
	DonationStatusFailed     = "FAILED"
)

type DonationRepo interface {
	Create(d *Donation) error
	GetDonations(d *Donation) ([]Donation, error)
	GetDonation(id int) (Donation, error)
	Update(d Donation) error
}

type Repo struct {
	database.Repo
}

func (r Repo) Create(d *Donation) error {
	var id int
	err := r.DB.Get(&id, `
	INSERT INTO donations (
		payment_id, 
		streamer_id, 
		amount, 
		message, 
		name, 
		status
	) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, d.PaymentID, d.StreamerID, d.Amount, d.Message, d.Name, d.Status)

	d.ID = id
	return err
}

func (r Repo) GetDonations(d *Donation) ([]Donation, error) {
	res := []Donation{}
	query := "SELECT * FROM donations"

	if d == nil || (d.PaymentID == "" && d.StreamerID == 0 && d.Status == "") {
		err := r.DB.Select(&res, query)
		return res, err
	}

	query += " WHERE"
	args := []any{}
	if d.PaymentID != "" {
		args = append(args, d.PaymentID)
		query += fmt.Sprintf(" payment_id = $%d", len(args))
	}

	if d.StreamerID != 0 {
		args = append(args, d.StreamerID)
		query += fmt.Sprintf(" streamer_id = $%d", len(args))
	}

	if d.Status != "" {
		args = append(args, d.Status)
		query += fmt.Sprintf(" status = $%d", len(args))
	}

	err := r.DB.Select(&res, query, args...)
	return res, err
}

func (r Repo) GetDonation(id int) (Donation, error) {
	var d Donation
	err := r.DB.Get(&d, "SELECT * FROM donations WHERE id = $1", id)
	return d, err
}

func (r Repo) Update(d Donation) error {
	_, err := r.DB.Exec(`
		UPDATE donations
		SET payment_id = $1, streamer_id = $2, amount = $3, message = $4, name = $5, status = $6
		WHERE id = $7`,
		d.PaymentID, d.StreamerID, d.Amount, d.Message, d.Name, d.Status, d.ID)
	return err
}
