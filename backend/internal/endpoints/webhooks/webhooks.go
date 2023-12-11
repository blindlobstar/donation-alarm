package webhooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/blindlobstar/donation-alarm/backend/internal/database/donation"
	"github.com/blindlobstar/donation-alarm/backend/internal/events"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

type WebhookConfig struct {
	Secret string
}
type WebhookEndpoint struct {
	DonationRepo donation.DonationRepo
	EventEmitter events.EventEmitter
	Config       WebhookConfig
}

func (we WebhookEndpoint) HandleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook error while parsing basic request. %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	signatureHeader := req.Header.Get("Stripe-Signature")
	event, err = webhook.ConstructEvent(payload, signatureHeader, we.Config.Secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "webhook signature verification failed. %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("successful payment for %d.", paymentIntent.Amount)
		donations, err := we.DonationRepo.GetDonations(&donation.Donation{PaymentID: paymentIntent.ID})
		if err != nil {
			log.Printf("error getting donation with PaymentID: %s\n", paymentIntent.ID)
			w.WriteHeader(http.StatusOK)
			return
		}
		if len(donations) < 1 {
			log.Printf("Donation not found. PaymentID: %s\n", paymentIntent.ID)
			w.WriteHeader(http.StatusOK)
			return
		}
		if len(donations) > 1 {
			log.Printf("more than one donation found. PaymentID: %s\n", paymentIntent.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if paymentIntent.Amount != int64(donations[0].Amount) {
			log.Printf("wrong amount. expected: %d, got: %d. PaymentID: %s\n", donations[0].Amount, paymentIntent.Amount, paymentIntent.ID)
		}

		donations[0].Status = donation.DonationStatusPayed
		if err = we.DonationRepo.Update(donations[0]); err != nil {
			log.Printf("can't update donation status after payment. DonationID: %d\n", donations[0].ID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = we.EventEmitter.Publish(events.DonationPayed{
			PaymentID:  donations[0].PaymentID,
			Message:    donations[0].Message,
			Name:       donations[0].Name,
			Status:     donations[0].Status,
			DonationID: donations[0].ID,
			StreamerID: donations[0].StreamerID,
			Amount:     donations[0].Amount,
		}, "DonationPayed"); err != nil {
			log.Printf("error publishing error. Err: %v", err)
		}

	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Then define and call a func to handle the successful attachment of a PaymentMethod.
		// handlePaymentMethodAttached(paymentMethod)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}
