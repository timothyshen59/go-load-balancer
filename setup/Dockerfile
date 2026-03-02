#Usage: cd .
# docker build -f setup/Dockerfile .

FROM golang:1.25-alpine AS builder
WORKDIR /app                    
COPY go.mod ./setup/servers.go ./
RUN go mod download && go build -o setup servers.go

FROM alpine:latest
WORKDIR /app                
COPY --from=builder /app/setup ./
CMD ["./setup"]


