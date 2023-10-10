package twitch_auth

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
	"github.com/nicklaw5/helix"
)

type Twitch struct {
	Client    *helix.Client
	Streamers database.StreamerRepo
}

type AuthRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (t *Twitch) Authenticate(w http.ResponseWriter, r *http.Request) error {
	var request AuthRequest
	json.NewDecoder(r.Body).Decode(&request)

	atr, err := t.Client.RequestUserAccessToken(request.Code)
	if err != nil {
		return err
	}
	if atr.StatusCode != http.StatusOK {
		log.Printf("error requesting twitch access token. statusCode: %d \nerror: %s \nerrorMessage: %s", atr.StatusCode, atr.Error, atr.ErrorMessage)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	_, vr, _ := t.Client.ValidateToken(atr.Data.AccessToken)
	if vr.StatusCode != http.StatusOK {
		log.Printf("error validating twitch access token. statusCode: %d \nerror: %s \nerrorMessage: %s", atr.StatusCode, atr.Error, atr.ErrorMessage)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	existingStreamers, err := t.Streamers.GetStreamers(database.Streamer{TwitchId: vr.Data.UserID})
	if err != nil {
		return err
	}

	// if user not exists, create one, publish event
	// and response with internal access and refresh tokens
	if len(existingStreamers) == 0 {
		streamer := database.Streamer{}
		streamer.SecretCode = generateSecretCode()
		streamer.TwitchId = vr.Data.UserID
		streamer.TwitchName = vr.Data.Login

		err := t.Streamers.CreateStreamer(&streamer)
		if err != nil {
			return err
		}
	} else if len(existingStreamers) > 1 {
		log.Printf("there is more than one user with same twitchId: %s \n", vr.Data.UserID)
		return errors.New("there is more than one user with same twitchId")
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

func generateSecretCode() string {
	cs := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	var rs strings.Builder
	for i := 0; i < 16; i++ {
		ri := rand.Intn(len(cs))
		rs.WriteByte(cs[ri])
	}
	return rs.String()
}
