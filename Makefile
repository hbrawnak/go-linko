BINARY=urlShotenerApp


up:
	@echo "Starting Docker images..."
	docker compose up -d
	@echo "Docker images started!"

down:
	@echo "Stopping docker compose..."
	docker compose down
	@echo "Done!"

up_build: build_app
	@echo "Stopping docker images (if running...)"
	docker compose down
	@echo "Building (when required) and starting docker images..."
	docker compose up --build -d
	@echo "Docker images built and started!"

build_app:
	@echo "Building app binary..."
	env GOOS=linux CGO_ENABLED=0 go build -o ${BINARY} ./cmd/api
	@echo "Done!"