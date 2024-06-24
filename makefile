EXEC := kiln-tezos-delegation

install:
	go get github.com/jackc/tern

build:
	go build -o dist/$(EXEC) main.go

.PHONY: build
run: build
	@./dist/$(EXEC)

clean:
	@rm -rf ./dist/
