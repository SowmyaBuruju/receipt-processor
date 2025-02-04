# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o receipt_processor

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/receipt_processor .
EXPOSE 8080
CMD ["./receipt_processor"]
