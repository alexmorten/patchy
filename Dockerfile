# Build frontend stage
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

# Copy frontend dependencies
COPY frontend/package.json frontend/package-lock.json* ./

# Install dependencies
RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build frontend
RUN npm run build

# Build backend stage
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download && go get golang.org/x/crypto/acme/autocert

# Copy source code
COPY . .

# Build the Go application
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates && mkdir -p /app/certs

WORKDIR /app

# Copy the compiled backend binary from the backend builder stage
COPY --from=backend-builder /server /app/server

# Create a directory for frontend assets
RUN mkdir -p /app/frontend/dist

# Copy frontend assets from the frontend builder stage
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Set environment variables that would normally come from your environment
ENV POSTGRES_CONNECTION_STRING="postgresql://postgres:postgres@postgres:5432/patchy"
ENV MEILISEARCH_URL="http://meilisearch:7700"
ENV PORT=7788
ENV HOST="0.0.0.0"
# ENV DOMAIN="yourdomain.com" # Uncomment and set this to enable autocert

# Expose ports
EXPOSE 7788
EXPOSE 80
EXPOSE 443

# Run the Go server
CMD ["/app/server"]