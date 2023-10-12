package donation

import (
	"encoding/json"
	"net/http"

	"github.com/blindlobstar/donation-alarm/backend/internal/database/donation"
	"github.com/blindlobstar/donation-alarm/backend/internal/database/streamer"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/paymentintent"
)

type Donation struct {
	DR donation.DonationRepo
	SR streamer.StreamerRepo
}

type CreateRequest struct {
	StreamerID int    `json:"streamerId"`
	Amount     int    `json:"amount"`
	Message    string `json:"message"`
	Name       string `json:"name"`
}

type CreateResponse struct {
	ClientSecret string `json:"clientSecret"`
}

func (de Donation) Create(w http.ResponseWriter, r *http.Request) error {
	var request CreateRequest
	json.NewDecoder(r.Body).Decode(&request)

	if request.StreamerID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	if request.Amount == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	streamer, err := de.SR.GetStreamerById(request.StreamerID)
	if err != nil {
		return err
	}

	if streamer == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	paymentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(request.Amount)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
	}
	pi, err := paymentintent.New(paymentParams)
	if err != nil {
		return err
	}
	donation := &donation.Donation{
		PaymentID:  pi.ID,
		StreamerID: request.StreamerID,
		Amount:     request.Amount,
		Message:    request.Message,
		Name:       request.Name,
		Status:     donation.DonationStatusCreated,
	}
	if err := de.DR.Create(donation); err != nil {
		return err
	}

	respBytes, err := json.Marshal(CreateResponse{
		ClientSecret: pi.ClientSecret,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
	return nil
}
