# ==========================================
# Multi-stage Dockerfile: WhatsApp Reminder
# ==========================================

# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend (Menggunakan Golang Alpine Terbaru)
FROM golang:alpine AS backend-builder
WORKDIR /app/backend

# Aktifkan auto toolchain agar dapat mendownload versi Go yang dibutuhkan secara otomatis
ENV GOTOOLCHAIN=auto

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/whatsapp-reminder ./cmd/server

# Stage 3: Minimal Runtime
FROM alpine:3.21 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S appgroup && adduser -S appuser -G appgroup \
    && mkdir -p /app/data /app/data/backups /app/data/logs \
    && chown -R appuser:appgroup /app

COPY --from=backend-builder --chown=appuser:appgroup /app/whatsapp-reminder /app/whatsapp-reminder
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist /app/frontend/dist

USER appuser

ENV PORT=8080 \
    DB_PATH=/app/data/reminders.db \
    WA_DB_PATH=/app/data/whatsapp.db \
    BACKUP_DIR=/app/data/backups \
    STATIC_DIR=/app/frontend/dist \
    DEFAULT_TIMEZONE=Asia/Jakarta

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/whatsapp-reminder"]
