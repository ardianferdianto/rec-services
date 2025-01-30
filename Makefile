APP_EXECUTABLE="bin/reconciliation-service"
APP_BOOTSTRAP="./cmd/server"

init:
	go install github.com/air-verse/air@latest
	go mod download
	go mod tidy
	cp ./config.yaml.tmpl ./config.yaml

deps-up:
	docker compose up -d

deps-down:
	docker compose down

compile:
	mkdir -p bin/
	go build -o $(APP_EXECUTABLE) $(APP_BOOTSTRAP)

migrate: compile
	$(APP_EXECUTABLE) migrate:run

partition: compile
	$(APP_EXECUTABLE) partition:create

run: compile
	env LOG_LEVEL=debug PUBSUB_EMULATOR_HOST=localhost:8681 ./bin/transaction-management start

air-http:
	air -c .dev/http.air.toml

copy-config:
	cp ./config.yaml.tmpl ./config.yaml