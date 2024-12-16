FROM golang:1.23.4-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o /app/bin/db-server ./db-server

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/db-server /app/db-server

EXPOSE 8080
ENTRYPOINT ["/app/db-server"] 