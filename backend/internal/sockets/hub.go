package sockets

import (
	"log"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients       map[int]*websocket.Conn
	connClientMap map[*websocket.Conn]int
	donationsC    chan DonationEvent
	registrationC chan RegistrationRequest
}

type RegistrationRequest struct {
	Conn       *websocket.Conn
	StreamerID int
}

type DonationEvent struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	Amount     int    `json:"amount"`
	StreamerID int    `json:"-"`
}

func CreateNew() Hub {
	return Hub{
		clients:       map[int]*websocket.Conn{},
		connClientMap: map[*websocket.Conn]int{},
		donationsC:    make(chan DonationEvent),
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

func (hub *Hub) Donate(donation DonationEvent) {
	hub.donationsC <- donation
}

func (hub *Hub) sendDonation(donation DonationEvent) {
	conn, ok := hub.clients[donation.StreamerID]
	if !ok {
		return
	}

	err := conn.WriteJSON(donation)
	if err != nil {
		log.Printf("can't send message through websocket. StreamerID: %d, Error: %v", donation.StreamerID, err)
		return
	}
}
