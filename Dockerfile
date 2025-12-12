# Build frontend
FROM node:18 as frontend-builder

WORKDIR /app
COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# Build backend
FROM golang:1.21 as backend-builder

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o gpx_grafana-reporter ./pkg

# Final image
FROM alpine:latest

WORKDIR /var/lib/grafana/plugins/progressio-grafanareporter-app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy built assets
COPY --from=frontend-builder /app/dist/ ./
COPY --from=backend-builder /app/gpx_grafana-reporter ./

# Create data directory
RUN mkdir -p data

# The plugin will be executed by Grafana
CMD ["/bin/sh"]
