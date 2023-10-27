package websockets

import (
	"log"
	"net/http"

	"github.com/blindlobstar/donation-alarm/backend/internal/database/streamer"
	"github.com/blindlobstar/donation-alarm/backend/internal/sockets"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WebSockets struct {
	StreamerRepo streamer.Repo
	Hub          *sockets.Hub
	Upgrader     websocket.Upgrader
}

func (ws WebSockets) Connect(w http.ResponseWriter, r *http.Request) error {
	secretCode := mux.Vars(r)["secretCode"]
	streamers, err := ws.StreamerRepo.GetStreamers(streamer.Streamer{SecretCode: secretCode})
	if err != nil {
		return err
	}

	if len(streamers) == 0 {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("Streamer not found. SecretCode: %s", secretCode)
		return nil
	}

	c, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	ws.Hub.RegisterClient(c, streamers[0].ID)
	return nil
}
