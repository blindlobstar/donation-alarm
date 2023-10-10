package backend

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
	"github.com/blindlobstar/donation-alarm/backend/internal/endpoints/twitch_auth"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nicklaw5/helix"
)

func main() {
	env := os.Getenv("BACKEND__ENV")
	if env == "" {
		env = "development"
	}

	if env == "development" {
		if err := godotenv.Load(); err != nil {
			log.Fatalln("error loading .env file")
		}
	}

	log.Println("connecting to database...")
	db, err := sqlx.Connect("postgres", os.Getenv("BACKEND__CONNECTION_STRING"))
	if err != nil {
		log.Fatalln("error connecting to database")
	}
	defer db.Close()

	rep := database.Repo{DB: db}
	log.Println("start database migration...")
	rep.Migrate()

	twitchClient, err := helix.NewClient(&helix.Options{
		ClientID:     os.Getenv("BACKEND__TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("BACKEND__TWITCH_CLIENT_SECRET"),
	})

	tw := twitch_auth.Twitch{
		Client:    twitchClient,
		Streamers: rep,
	}

	r := mux.NewRouter()
	r.HandleFunc("/auth/twitch", errorHandler(tw.Authenticate)).Methods("POST")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	log.Println("server starting...")
	s := http.Server{
		Addr:    ":80",
		Handler: r,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-sig
	log.Println("shouting down server...")
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
