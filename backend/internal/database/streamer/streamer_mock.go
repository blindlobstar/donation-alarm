package streamer

type StreamerMock struct {
	Streamers []Streamer
}

func (sm *StreamerMock) CreateStreamer(s *Streamer) error {
	s.ID = len(sm.Streamers)
	sm.Streamers = append(sm.Streamers, *s)

	return nil
}

func (sm *StreamerMock) GetStreamers(s Streamer) ([]Streamer, error) {

	if s.TwitchId == "" && s.TwitchName == "" {
		res := make([]Streamer, len(sm.Streamers))
		copy(res, sm.Streamers)
		return res, nil
	}

	res := make([]Streamer, 0, len(sm.Streamers))
	if s.TwitchId != "" {
		for _, streamer := range sm.Streamers {
			if streamer.TwitchId == s.TwitchId {
				res = append(res, streamer)
			}
		}
	}

	if s.TwitchName != "" {
		for _, streamer := range sm.Streamers {
			if streamer.TwitchName == s.TwitchName {
				res = append(res, streamer)
			}
		}
	}

	return res, nil
}

func (sm *StreamerMock) GetStreamerById(id int) (*Streamer, error) {
	for _, s := range sm.Streamers {
		if s.ID == id {
			return &Streamer{
				ID:         id,
				TwitchId:   s.TwitchId,
				TwitchName: s.TwitchName,
				SecretCode: s.SecretCode,
			}, nil
		}
	}

	return nil, nil
}
