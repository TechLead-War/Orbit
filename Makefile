BINARY_NAME=orbit
DB_URL=postgres://postgres:password@localhost:5432/Orbit?sslmode=disable

build:
	go build -o ${BINARY_NAME} cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

clean:
	go clean
	rm -f ${BINARY_NAME}

deps:
	go mod download
	go mod tidy

migrate-up:
	migrate -database ${DB_URL} -path migrations up

migrate-down:
	migrate -database ${DB_URL} -path migrations down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

.PHONY: build run test clean deps migrate-up migrate-down migrate-create docker-up docker-down 