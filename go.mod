module planets-server

go 1.25.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/rs/cors v1.11.1
	golang.org/x/oauth2 v0.30.0
	golang.org/x/time v0.12.0
)

require cloud.google.com/go/compute/metadata v0.3.0 // indirect
