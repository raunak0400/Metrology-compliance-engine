# syntax=docker/dockerfile:1
# ---- Build ----
FROM golang:1.23-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 go build -o /out/seed ./cmd/seed

# ---- Production runtime ----
FROM alpine:3.20 AS runtime
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/api /out/worker /out/seed ./
COPY migrations ./migrations
ENV GO_ENV=production
EXPOSE 4000
# Default command runs the API. Override to `./worker` for the worker fleet or
# `./seed` to run the one-off database seed.
CMD ["./api"]
