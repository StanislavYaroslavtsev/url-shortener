FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY services/stats/ ./services/stats/
COPY services/url-shortener/go.mod services/url-shortener/go.sum ./services/url-shortener/

WORKDIR /app/services/stats

RUN go build -o ./bin/app ./cmd/server

FROM builder AS development
RUN go install github.com/air-verse/air@latest
EXPOSE 50051
CMD ["air", "-c", ".air.toml"]