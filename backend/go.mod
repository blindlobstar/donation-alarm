module github.com/blindlobstar/donation-alarm/backend

go 1.20

require (
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/nicklaw5/helix v1.25.0
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/stripe/stripe-go/v75 v75.10.0
)

require (
	github.com/golang-jwt/jwt v3.2.1+incompatible // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)
