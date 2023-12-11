package handlers

import (
	"github.com/blindlobstar/donation-alarm/backend/internal/events"
	"github.com/blindlobstar/donation-alarm/backend/internal/sockets"
)

type DonationPayedHandler struct {
	hub *sockets.Hub
}

func NewDonationPayedHandler(hub *sockets.Hub) DonationPayedHandler {
	return DonationPayedHandler{
		hub: hub,
	}
}

func (h DonationPayedHandler) Handle(event any) error {
	dpe := event.(events.DonationPayed)
	h.hub.Donate(sockets.DonationEvent{
		Name:       dpe.Name,
		Text:       dpe.Message,
		Amount:     dpe.Amount / 100,
		StreamerID: dpe.StreamerID,
	})
	return nil
}
