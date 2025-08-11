# syntax=docker/dockerfile:1
# ─────────────────────────────────────────────
# Stage 1: builder – download modules & compile
FROM golang:1.23.5-alpine AS builder
LABEL stage=builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/server ./cmd/server

# ─────────────────────────────────────────────
FROM golang:1.23.5-alpine AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/server ./server
COPY .env ./.env
COPY internal/config/certs ./internal/config/certs
COPY migrations ./migrations

# Copy web directory for UI templates and static files
COPY web ./web

EXPOSE 8080

ENV APP_ENVIRONMENT=local
CMD ["./server"]