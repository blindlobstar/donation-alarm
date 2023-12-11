package events

type DonationPayed struct {
	PaymentID  string
	Message    string
	Name       string
	Status     string
	DonationID int
	StreamerID int
	Amount     int
}
