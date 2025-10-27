FROM golang:latest AS builder

WORKDIR /wordbuddy

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg

RUN go build -o ./.bin/bot ./cmd/bot/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates libssl3 && rm -rf /var/lib/apt/lists/*

WORKDIR /wordbuddy
COPY --from=builder /wordbuddy/.bin/bot .

CMD ["./bot"]