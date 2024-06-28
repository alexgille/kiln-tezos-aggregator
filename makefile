EXEC := kiln-tezos-delegation
TERN := $(shell go env GOPATH)/bin/tern

install:
	go install github.com/jackc/tern/v2@latest

docker-up:
	docker compose -f build/docker-compose.yaml up -d

docker-down:
	docker compose -f build/docker-compose.yaml down

migrate-up:
	$(TERN) -c repository/tern/tern.conf migrate --migrations repository/migrations/

migrate-down:
	$(TERN) -c repository/tern/tern.conf migrate --migrations repository/migrations/ --destination -1

build:
	go build -o dist/$(EXEC) main.go

test:
	go test -timeout 10s ./...

.PHONY: build
run: build
	@./dist/$(EXEC)

clean:
	@rm -rf ./dist/
