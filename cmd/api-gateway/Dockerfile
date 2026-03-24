FROM golang:1.26.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api-gateway ./cmd/api-gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/api-gateway /app/api-gateway
COPY --from=builder /app/web /app/web
CMD ["/app/api-gateway"]
