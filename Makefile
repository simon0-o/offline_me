.PHONY: build run frontend-install frontend-build frontend-dev docker-build docker-run clean

# Build the Go binary from backend directory
build:
	cd backend && go build -o ../offline_me ./cmd/server

# Install frontend dependencies
frontend-install:
	cd frontend && npm install

# Build frontend for production
frontend-build:
	cd frontend && npm run build

# Run frontend development server
frontend-dev:
	cd frontend && npm run dev

frontend-run: frontend-build
	cd frontend && npm run start

# Build both frontend and backend
build-all: frontend-build build

# Run the application
run:
	cd backend && go run ./cmd/server

# Build Docker image
docker-build:
	docker build -t offline_me .

# Run Docker container
docker-run:
	docker run -p 8080:8080 -v $(PWD)/worktime.db:/root/worktime.db offline_me
