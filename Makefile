SHA = $(shell git rev-parse HEAD)
LDFLAGS = -X 'github.com/lukeshay/records/pkg/config.Version=$(SHA)'
TMP = ./tmp
BIN = $(TMP)/bin/records
CERTS = $(TMP)/certs

.DEFAULT: build

clean-bin:
	@rm -rf $(shell dirname $(BIN))

dev: certs
ifndef $(shell command -v dot 2> /dev/null)
	@go install github.com/cosmtrek/air@latest
endif
	@air -c .air.toml

build: clean-bin
	bun run build
	go build -ldflags="$(LDFLAGS)" -o $(BIN)

run: build certs
	$(BIN)

kill:
	@kill -9 $(shell lsof -t -i:8080)

certs:
	@ mkdir -p $(CERTS)
	@openssl genrsa -out $(CERTS)/server.key 2048
	@openssl req -new -x509 -key $(CERTS)/server.key -out $(CERTS)/server.pem -days 365 -subj "/C=US/ST=Iowa/L=West Des Moines/O=Luke Shay/OU=core/CN=Luke Shay/emailAddress=$(shell whoami)@email.com"

