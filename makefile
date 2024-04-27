# Include variables from the .envrc file
include .envrc

# Create  the new confirm target
confirm:
	@echo -n "Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]

run/api:
	go run ./cmd/api

db/psql:
	psql ${LOGOCOMP}

db/migrations/new:
	@echo "Creating migration files for ${name}..."
	migrate create -seq -ext=.sql -dir=./migrations ${name}

db/migrations/up:
	@echo "Running up migrations..."
	migrate -path ./migrations -database ${LOGOCOMP} up

build/api:
	@echo "Building cmd/api..."
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api	
