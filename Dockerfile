# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /main .
COPY templates ./templates
COPY static ./static

EXPOSE 8080

CMD ["/app/main"]
