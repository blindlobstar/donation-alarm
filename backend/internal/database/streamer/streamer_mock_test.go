//go:build unit
// +build unit

package streamer

import (
	"testing"
)

func TestCreateStreamer(t *testing.T) {
	sm := &StreamerMock{}
	streamer := &Streamer{
		TwitchId:   "twitch123",
		TwitchName: "teststreamer",
		SecretCode: "secret123",
	}

	err := sm.CreateStreamer(streamer)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if streamer.ID != 0 {
		t.Errorf("Expected ID to be 0, got %d", streamer.ID)
	}

	if len(sm.Streamers) != 1 {
		t.Errorf("Expected Streamers slice length to be 1, got %d", len(sm.Streamers))
	}

	if sm.Streamers[0] != *streamer {
		t.Errorf("Expected the stored Streamer to be equal to the created Streamer")
	}
}

func TestGetStreamers(t *testing.T) {
	sm := &StreamerMock{}
	streamer1 := Streamer{
		TwitchId:   "twitch123",
		TwitchName: "teststreamer1",
		SecretCode: "secret123",
	}
	streamer2 := Streamer{
		TwitchId:   "twitch456",
		TwitchName: "teststreamer2",
		SecretCode: "secret456",
	}
	sm.Streamers = []Streamer{streamer1, streamer2}

	// Test case 1: Get all streamers
	allStreamers, err := sm.GetStreamers(Streamer{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(allStreamers) != 2 {
		t.Errorf("Expected 2 streamers, got %d", len(allStreamers))
	}

	// Test case 2: Get streamer by Twitch ID
	byTwitchID, err := sm.GetStreamers(Streamer{TwitchId: "twitch123"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(byTwitchID) != 1 || byTwitchID[0] != streamer1 {
		t.Errorf("Expected 1 streamer with Twitch ID 'twitch123', got %+v", byTwitchID)
	}

	// Test case 3: Get streamer by Twitch Name
	byTwitchName, err := sm.GetStreamers(Streamer{TwitchName: "teststreamer2"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(byTwitchName) != 1 || byTwitchName[0] != streamer2 {
		t.Errorf("Expected 1 streamer with Twitch Name 'teststreamer2', got %+v", byTwitchName)
	}
}

func TestGetStreamerById(t *testing.T) {
	sm := &StreamerMock{}
	streamer1 := Streamer{
		ID:         0,
		TwitchId:   "twitch123",
		TwitchName: "teststreamer1",
		SecretCode: "secret123",
	}
	sm.Streamers = []Streamer{streamer1}

	// Test case 1: Get a streamer by ID
	retrievedStreamer, err := sm.GetStreamerById(0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedStreamer == nil {
		t.Errorf("Expected a valid streamer, got nil")
	} else if *retrievedStreamer != streamer1 {
		t.Errorf("Expected the retrieved streamer to be %+v, got %+v", streamer1, *retrievedStreamer)
	}

	// Test case 2: Get a non-existent streamer by ID
	retrievedStreamer, err = sm.GetStreamerById(1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrievedStreamer != nil {
		t.Errorf("Expected nil for a non-existent streamer, got %+v", *retrievedStreamer)
	}
}
