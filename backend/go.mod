module github.com/taskflow/backend

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.12.3
	golang.org/x/crypto v0.45.0
)

replace golang.org/x/crypto => /Users/ajay.patel/Library/CloudStorage/OneDrive-UKG/Desktop/TaskFlow/backend/vendor_local/golang-crypto

replace golang.org/x/sys => /Users/ajay.patel/Library/CloudStorage/OneDrive-UKG/Desktop/TaskFlow/backend/vendor_local/golang-sys

replace golang.org/x/net => /Users/ajay.patel/Library/CloudStorage/OneDrive-UKG/Desktop/TaskFlow/backend/vendor_local/golang-net

replace golang.org/x/term => /Users/ajay.patel/Library/CloudStorage/OneDrive-UKG/Desktop/TaskFlow/backend/vendor_local/golang-term

replace golang.org/x/text => /Users/ajay.patel/Library/CloudStorage/OneDrive-UKG/Desktop/TaskFlow/backend/vendor_local/golang-text
