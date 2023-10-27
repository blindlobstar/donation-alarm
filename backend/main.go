package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/blindlobstar/donation-alarm/backend/internal/database"
	"github.com/blindlobstar/donation-alarm/backend/internal/database/donation"
	"github.com/blindlobstar/donation-alarm/backend/internal/database/streamer"
	donationendpoint "github.com/blindlobstar/donation-alarm/backend/internal/endpoints/donation"
	"github.com/blindlobstar/donation-alarm/backend/internal/endpoints/twitch_auth"
	"github.com/blindlobstar/donation-alarm/backend/internal/endpoints/webhooks"
	"github.com/blindlobstar/donation-alarm/backend/internal/endpoints/websockets"
	"github.com/blindlobstar/donation-alarm/backend/internal/sockets"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nicklaw5/helix"
	"github.com/stripe/stripe-go/v75"
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
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	rep := database.Repo{DB: db}
	log.Println("start database migration...")
	rep.Migrate()

	twitchClient, err := helix.NewClient(&helix.Options{
		ClientID:     os.Getenv("BACKEND__TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("BACKEND__TWITCH_CLIENT_SECRET"),
		RedirectURI:  "http://localhost",
	})

	tw := twitch_auth.Twitch{
		Client:    twitchClient,
		Streamers: streamer.Repo{Repo: rep},
	}

	log.Println(os.Getenv("STRIPE_API_KEY"))
	stripe.Key = os.Getenv("STRIPE_API_KEY")
	de := donationendpoint.Donation{
		DR: donation.Repo{Repo: rep},
		SR: streamer.Repo{Repo: rep},
	}

	hub := sockets.CreateNew()
	go hub.Run()

	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws := websockets.WebSockets{
		StreamerRepo: streamer.Repo{Repo: rep},
		Hub:          &hub,
		Upgrader:     upgrader,
	}

	webhook := webhooks.WebhookEndpoint{
		WebsocketHub: &hub,
		DonationRepo: donation.Repo{Repo: rep},
	}
	r := mux.NewRouter()
	r.HandleFunc("/auth/twitch", errorHandler(tw.Authenticate)).Methods(http.MethodPost)
	r.HandleFunc("/donation", useCORS(errorHandler(de.Create))).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/ws/{secretCode}", errorHandler(ws.Connect))
	r.HandleFunc("/webhooks", webhook.HandleWebhook).Methods(http.MethodPost)

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

func useCORS(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		f(w, r)
	}
}

func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
