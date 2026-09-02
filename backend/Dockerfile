# syntax=docker/dockerfile:1
# ---- Build ----
FROM golang:1.26-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 go build -o /out/seed ./cmd/seed
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

# ---- Production runtime ----
FROM alpine:3.20 AS runtime
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/api /out/worker /out/seed /out/migrate ./
COPY migrations ./migrations
ENV GO_ENV=production
EXPOSE 4000
# Default command runs the API. Override to `./worker` for the worker fleet,
# `./migrate up` to apply pending migrations, or `./seed` for the one-off
# database seed.
CMD ["./api"]
