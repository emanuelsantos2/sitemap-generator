# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache gcc musl-dev
RUN go build -o sitemap-generator .

# Final stage
FROM alpine:latest
RUN apk add --no-cache sqlite
WORKDIR /app
COPY --from=builder /app/sitemap-generator .
EXPOSE 3000
CMD ["./sitemap-generator"]