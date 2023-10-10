package database

type Streamer struct {
	ID         int
	TwitchId   string `db:"twitch_id"`
	TwitchName string `db:"twitch_name"`
	SecretCode string `db:"secret_code"`
}

type StreamerRepo interface {
	CreateStreamer(s *Streamer) error
	GetStreamers(s Streamer) ([]Streamer, error)
	GetStreamerById(id int) (*Streamer, error)
}

func (r Repo) CreateStreamer(s *Streamer) error {
	_, err := r.DB.Exec("INSERT INTO streamers (twitch_name, secret_code) VALUES ($1, $2)", s.TwitchName, s.SecretCode)
	return err
}

func (r Repo) GetStreamers(s Streamer) ([]Streamer, error) {
	res := []Streamer{}
	req := "SELECT * FROM streamer"
	if s.TwitchId == "" && s.TwitchName == "" {
		err := r.DB.Select(res, req)
		return res, err
	}

	req += " WHERE"
	args := []any{}

	if s.TwitchId != "" {
		req += " twitch_id = ?"
		args = append(args, s.TwitchId)
	}

	if s.TwitchName != "" {
		req += " twitch_name = ?"
		args = append(args, s.TwitchName)
	}

	err := r.DB.Select(res, req, args...)
	return res, err
}

func (r Repo) GetStreamerById(id int) (*Streamer, error) {
	res := &Streamer{}
	err := r.DB.Get(res, "SELECT * FROM streamer WHERE id = $1", id)

	return res, err
}
