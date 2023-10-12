package streamer

import (
	"fmt"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
)

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

type Repo struct {
	database.Repo
}

func (r Repo) CreateStreamer(s *Streamer) error {
	_, err := r.DB.Exec("INSERT INTO streamers (twitch_id, twitch_name, secret_code) VALUES ($1, $2, $3)", s.TwitchId, s.TwitchName, s.SecretCode)
	return err
}

func (r Repo) GetStreamers(s Streamer) ([]Streamer, error) {
	res := []Streamer{}
	req := "SELECT * FROM streamers"
	if s.TwitchId == "" && s.TwitchName == "" {
		err := r.DB.Select(res, req)
		return res, err
	}

	req += " WHERE"
	args := []any{}

	if s.TwitchId != "" {
		args = append(args, s.TwitchId)
		req += fmt.Sprintf(" twitch_id = $%d", len(args))
	}

	if s.TwitchName != "" {
		args = append(args, s.TwitchName)
		req += fmt.Sprintf(" twitch_name = $%d", len(args))
	}

	err := r.DB.Select(&res, req, args...)
	return res, err
}

func (r Repo) GetStreamerById(id int) (*Streamer, error) {
	res := &Streamer{}
	err := r.DB.Get(res, "SELECT * FROM streamers WHERE id = $1", id)

	return res, err
}
