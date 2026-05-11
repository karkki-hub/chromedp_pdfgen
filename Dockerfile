# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# Install Chromium + required libs
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ca-certificates \
    ttf-freefont

# Set Chromium environment variables
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROMIUM_PATH=/usr/bin/chromium-browser

# Copy binary
COPY --from=builder /app/app .

# Expose app port
EXPOSE 8080

# Start app
CMD ["./app"]