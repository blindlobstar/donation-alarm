package twitch_auth

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/blindlobstar/donation-alarm/backend/internal/database/streamer"
	"github.com/gorilla/sessions"
	"github.com/nicklaw5/helix"
)

const (
	stateCallbackKey = "oauth-state-callback"
	oauthSessionName = "oauth-session"
	oauthTokenKey    = "oauth-token"
)

type Twitch struct {
	Client      *helix.Client
	Streamers   streamer.StreamerRepo
	CookieStore *sessions.CookieStore
}

type AuthRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func (t *Twitch) HandleLogin(w http.ResponseWriter, r *http.Request) error {
	session, err := t.CookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
	}

	var tokenBytes [255]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return err
	}

	state := hex.EncodeToString(tokenBytes[:])

	session.AddFlash(state, stateCallbackKey)

	if err := session.Save(r, w); err != nil {
		return err
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", "1a0tcerhh4x1qpijw1ukc8hqbf415o")
	params.Add("redirect_uri", "http://localhost:8888/auth/twitch")
	params.Add("scope", "channel:manage:polls channel:read:polls")
	params.Add("state", state)

	authCodeUrl := "https://id.twitch.tv/oauth2/authorize?" + params.Encode()
	http.Redirect(w, r, authCodeUrl, http.StatusTemporaryRedirect)

	return err
}

// HandleOauth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func (t *Twitch) HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := t.CookieStore.Get(r, oauthSessionName)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	switch stateChallenge, state := session.Flashes(stateCallbackKey), r.FormValue("state"); {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'", state, stateChallenge[0])
	}

	// saving session after reading Flashes
	if err := session.Save(r, w); err != nil {
		log.Printf("error saving session: %s", err)
		return err
	}

	if err != nil {
		return err
	}

	atr, err := t.Client.RequestUserAccessToken(r.FormValue("code"))
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
	existingStreamers, err := t.Streamers.GetStreamers(streamer.Streamer{TwitchId: vr.Data.UserID})
	if err != nil {
		return err
	}

	if len(existingStreamers) > 1 {
		log.Printf("there is more than one user with same twitchId: %s \n", vr.Data.UserID)
		return errors.New("there is more than one user with same twitchId")
	}

	var authStreamer streamer.Streamer
	// if user not exists, create one, publish event
	if len(existingStreamers) == 0 {
		authStreamer = streamer.Streamer{}
		authStreamer.SecretCode = generateSecretCode()
		authStreamer.TwitchId = vr.Data.UserID
		authStreamer.TwitchName = vr.Data.Login

		err := t.Streamers.CreateStreamer(&authStreamer)
		if err != nil {
			return err
		}
	} else {
		authStreamer = existingStreamers[0]
	}

	// add the oauth token to session
	session.Values[oauthTokenKey] = authStreamer.ID
	if err = sessions.Save(r, w); err != nil {
		log.Println("all good, redirecting with session")
		return err
	}

	http.Redirect(w, r, "http://localhost:5173/", http.StatusTemporaryRedirect)
	return
}

func (t *Twitch) Authenticate(w http.ResponseWriter, r *http.Request) error {
	var request AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return err
	}

	atr, err := t.Client.RequestUserAccessToken(request.Code)
	if err != nil {
		return err
	}
	if atr.StatusCode != http.StatusOK {
		log.Printf("error requesting twitch access token. statusCode: %d \nerror: %s \nerrorMessage: %s\n", atr.StatusCode, atr.Error, atr.ErrorMessage)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	_, vr, _ := t.Client.ValidateToken(atr.Data.AccessToken)
	if vr.StatusCode != http.StatusOK {
		log.Printf("error validating twitch access token. statusCode: %d \nerror: %s \nerrorMessage: %s\n", atr.StatusCode, atr.Error, atr.ErrorMessage)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	existingStreamers, err := t.Streamers.GetStreamers(streamer.Streamer{TwitchId: vr.Data.UserID})
	if err != nil {
		return err
	}

	// if user not exists, create one, publish event
	// and response with internal access and refresh tokens
	if len(existingStreamers) == 0 {
		streamer := streamer.Streamer{}
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
