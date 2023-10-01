package twitch_auth

import (
	"donation-alarm/internal/backend/database"
	"encoding/json"
	"log"
	"net/http"

	"github.com/nicklaw5/helix"
)

type Twitch struct {
	Client *helix.Client
	DB     database.Repo
}

type AuthRequest struct {
	Code  string
	State string
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
	existingStreamers, err := t.DB.GetStreamers(database.Streamer{TwitchId: vr.Data.UserID})
	if err != nil {
		return err
	}

	// if user not exists, create one, publish event
	// and response with internal access and refresh tokens
	if len(existingStreamers) == 0 {
		err := t.DB.CreateStreamer(&existingStreamers[0])
		if err != nil {
			return err
		}
		// TODO: publish event

	} else if len(existingStreamers) > 1 {
		log.Printf("there is more than one user with same twitchId: %s \n", vr.Data.UserID)
		return err
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}