FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/autodidact-api ./cmd/api

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /data

COPY --from=builder /app/autodidact-api /app/autodidact-api

ENV PORT=8080
ENV HOST=0.0.0.0
ENV DB_PATH=/data/autodidact.db
ENV AUTH_TOKEN=autodidact-local-key

EXPOSE 8080

VOLUME ["/data"]

CMD ["/app/autodidact-api"]
