# 1 
FROM golang:alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o mstf-app ./cmd/app/main.go

# 2
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/mstf-app .

EXPOSE 8080
CMD ["./mstf-app"]
