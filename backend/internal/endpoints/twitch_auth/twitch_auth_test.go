package twitch_auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
	"github.com/nicklaw5/helix"
)

type mockHTTPClient struct {
	mockHandler http.HandlerFunc
}

func (mtc *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mtc.mockHandler)
	handler.ServeHTTP(rr, req)

	return rr.Result(), nil
}

func TestAuthenticate(t *testing.T) {
	streamerMock := &database.StreamerMock{
		Streamers: []database.Streamer{},
	}
	twitchAuth := Twitch{
		Client:    &helix.Client{},
		Streamers: streamerMock,
	}

	// Test case 1: Got error while trying to get twitch access token
	twitchAuth.Client, _ = helix.NewClient(&helix.Options{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		HTTPClient: &mockHTTPClient{
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"status":400,"message":"Invalid authorization code"}`))
			},
		},
	})
	req := httptest.NewRequest("POST", "/auth/twitch", strings.NewReader(`{"code": "abc", "state": "state"}`))
	rr := httptest.NewRecorder()
	err := twitchAuth.Authenticate(rr, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got: %d", rr.Result().StatusCode)
	}

	// Test case 2
	var counter int
	twitchAuth.Client, _ = helix.NewClient(&helix.Options{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		HTTPClient: &mockHTTPClient{
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				counter++
				if r.Method == "POST" && r.URL.Path == "/oauth2/token" {
					if r.URL.Query().Get("code") != "valid-code" {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(`{"status":400,"message":"Invalid authorization code"}`))
						return
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"access_token":"invalid-access-token","expires_in":14154,"refresh_token":"fiuhgaofohofhohdflhoiwephvlhowiehfoi","scope":["analytics:read:games","bits:read","clips:edit","user:edit","user:read:email"]}`))
					return
				}

				if r.Method == "GET" && r.URL.Path == "/oauth2/validate" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"status":401,"message":"invalid access token"}`))
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
		},
	})
	req = httptest.NewRequest("POST", "/auth/twitch", strings.NewReader(`{"code": "valid-code", "state": "state"}`))
	rr = httptest.NewRecorder()
	err = twitchAuth.Authenticate(rr, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if counter != 2 {
		t.Fatalf("expected 2 twitch request, got: %d", counter)
	}
	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got: %d", rr.Result().StatusCode)
	}

	// Test case 3
	twitchAuth.Client, _ = helix.NewClient(&helix.Options{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		HTTPClient: &mockHTTPClient{
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" && r.URL.Path == "/oauth2/token" {
					if r.URL.Query().Get("code") != "valid-code" {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(`{"status":400,"message":"Invalid authorization code"}`))
						return
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"access_token":"invalid-access-token","expires_in":14154,"refresh_token":"fiuhgaofohofhohdflhoiwephvlhowiehfoi","scope":["analytics:read:games","bits:read","clips:edit","user:edit","user:read:email"]}`))
					return
				}

				if r.Method == "GET" && r.URL.Path == "/validate" {
					w.WriteHeader(http.StatusUnauthorized)
					if authToken, ok := r.Header["Authorization"]; ok && len(authToken) > 0 && authToken[0] == "Bearer valid-access-token" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"client_id":"leadku246lkasdj6l6ljsd2","login":"authduser","scopes":["user:read:email"],"user_id":"12345","expires_in":5243778}`))
					}
					return
				}
			},
		},
	})
	req = httptest.NewRequest("POST", "/auth/twitch", strings.NewReader(`{"code": "valid-code", "state": "state"}`))
	rr = httptest.NewRecorder()
	err = twitchAuth.Authenticate(rr, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rr.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got: %d", rr.Result().StatusCode)
	}
	if len(streamerMock.Streamers) != 1 {
		t.Fatalf("expected 1 streamer, got %d", len(streamerMock.Streamers))
	}

	// Test case 4: auth same user
	req = httptest.NewRequest("POST", "/auth/twitch", strings.NewReader(`{"code": "valid-code", "state": "state"}`))
	rr = httptest.NewRecorder()
	err = twitchAuth.Authenticate(rr, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rr.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got: %d", rr.Result().StatusCode)
	}
	if len(streamerMock.Streamers) != 1 {
		t.Fatalf("expected 1 streamer, got %d", len(streamerMock.Streamers))
	}

	// Test case 5: same users in db
	streamerMock.Streamers = append(streamerMock.Streamers, database.Streamer{TwitchId: "12345"})
	req = httptest.NewRequest("POST", "/auth/twitch", strings.NewReader(`{"code": "valid-code", "state": "state"}`))
	rr = httptest.NewRecorder()
	err = twitchAuth.Authenticate(rr, req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "there is more than one user with same twitchId" {
		t.Fatalf("expected \"there is more than one user with same twitchId\", got: %s", err.Error())
	}
}
