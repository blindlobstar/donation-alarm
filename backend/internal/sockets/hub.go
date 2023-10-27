package sockets

import (
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients       map[int]*websocket.Conn
	connClientMap map[*websocket.Conn]int
	donationsC    chan Donation
	registrationC chan RegistrationRequest
}

type RegistrationRequest struct {
	Conn       *websocket.Conn
	StreamerID int
}

type Donation struct {
	Amount     int
	Text       string
	StreamerID int
}

type DonationEvent struct {
	Amount int    `json:"amount"`
	Text   string `json:"text"`
}

func CreateNew() Hub {
	return Hub{
		clients:       map[int]*websocket.Conn{},
		connClientMap: map[*websocket.Conn]int{},
		donationsC:    make(chan Donation),
		registrationC: make(chan RegistrationRequest),
	}
}

func (hub *Hub) RegisterClient(conn *websocket.Conn, streamerID int) {
	hub.registrationC <- RegistrationRequest{
		Conn:       conn,
		StreamerID: streamerID,
	}
}

func (hub *Hub) Run() {
	for {
		select {
		case rr := <-hub.registrationC:
			hub.clients[rr.StreamerID] = rr.Conn
			hub.connClientMap[rr.Conn] = rr.StreamerID
		case d := <-hub.donationsC:
			hub.sendDonation(d)
		}
	}
}

func (hub *Hub) Donate(donation Donation) {
	hub.donationsC <- donation
}

func (hub *Hub) sendDonation(donation Donation) {
	conn, ok := hub.clients[donation.StreamerID]
	if !ok {
		return
	}

	err := conn.WriteJSON(DonationEvent{Amount: donation.Amount, Text: donation.Text})
	if err != nil {
		return
	}
}
