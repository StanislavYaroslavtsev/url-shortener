FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY services/url-shortener/ ./services/url-shortener/
COPY services/stats/go.mod services/stats/go.sum ./services/stats/

WORKDIR /app/services/url-shortener

RUN go build -o ./bin/app ./cmd/server

FROM builder AS development
RUN go install github.com/air-verse/air@latest
EXPOSE 3000
CMD ["air", "-c", ".air.toml"]