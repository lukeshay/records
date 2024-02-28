FROM oven/bun:1 as assets

WORKDIR /app

COPY package.json bun.lockb ./

RUN bun install

COPY tailwind.config.js globals.css ./
COPY assets ./assets
COPY templates ./templates

RUN bun run build

FROM golang:1.22.0-alpine3.19 as builder

WORKDIR /go/src/github.com/lukeshay/records

RUN apk update && apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY pkg ./pkg
COPY templates ./templates
COPY assets ./assets
COPY main.go ./
COPY .git ./.git

RUN CGO_ENABLED=0 GOOS=linux go build  -ldflags="-X 'github.com/lukeshay/pkg/config/config.Version=$(git rev-parse HEAD)'" -o /go/bin/records

FROM scratch

COPY --from=builder /go/bin/records /go/bin/records

ENTRYPOINT ["/go/bin/records"]

