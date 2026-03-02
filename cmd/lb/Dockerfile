#docker build -f cmd/lb/Dockerfile .
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY ../../go.mod ../../go.sum ./
RUN go mod download

COPY ../../ ./
RUN go build -o lb ./cmd/lb/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lb .
CMD ["./lb"]







