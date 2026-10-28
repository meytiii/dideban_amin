FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/herasat .

FROM alpine:3.20

RUN apk add --no-cache \
    python3 \
    py3-pip \
    ca-certificates \
    tzdata \
    postgresql16-client && \
    pip install --no-cache-dir --break-system-packages python-docx==1.1.2

RUN addgroup -g 10001 -S herasat && \
    adduser -u 10001 -S herasat -G herasat

WORKDIR /app

COPY --from=builder --chown=herasat:herasat /build/herasat /app/herasat
COPY --chown=herasat:herasat frontend/ /app/frontend/
COPY --chown=herasat:herasat templates/ /app/templates/
COPY --chown=herasat:herasat scripts/ /app/scripts/
COPY --chown=herasat:herasat schema.sql /app/schema.sql

RUN mkdir -p /app/secure_uploads /app/backups && \
    chown -R herasat:herasat /app/secure_uploads /app/backups

USER herasat

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health | grep -q "secure_system_active" || exit 1

ENTRYPOINT ["/app/herasat"]
