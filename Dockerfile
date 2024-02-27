FROM golang:1.22.0-alpine3.19 as builder

WORKDIR /go/src/github.com/lukeshay/records

RUN apk update && apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY pkg ./pkg
COPY main.go ./
COPY .git ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/records

FROM scratch

COPY --from=builder /go/bin/records /go/bin/records

ENTRYPOINT ["/go/bin/records"]

