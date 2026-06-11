# ---------------------------------------------------------------------------
# Root Dockerfile for single-app deployment (Render or any Docker host).
#
# Builds the Go API and the Next.js frontend into one image. At runtime the
# Go API listens on 127.0.0.1:8090 and the Next server (platform PORT,
# default 3000) proxies /api/* and /healthz to it, so the whole app is served
# from a single origin and no CORS configuration is needed.
#
# For local development prefer `docker compose up --build`, which runs the
# services separately.
# ---------------------------------------------------------------------------

# --- Stage 1: build the Go API ---
FROM golang:1.26-alpine AS backend-build
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

# --- Stage 2: build the Next.js frontend ---
FROM node:22-alpine AS frontend-build
WORKDIR /src

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
# Empty NEXT_PUBLIC_API_URL => the browser uses relative URLs, which the
# Next server rewrites to the in-container Go API (see next.config.ts).
ENV NEXT_PUBLIC_API_URL=""
ENV API_PROXY_TARGET="http://127.0.0.1:8090"
RUN npm run build

# --- Stage 3: runtime ---
FROM node:22-alpine
WORKDIR /app
# PORT is intentionally not set here: the platform (e.g. Render) provides it
# for the public Next.js server, while start.sh pins the Go API to 8090.
ENV NODE_ENV=production \
    UPLOAD_DIR=/data/uploads \
    HOSTNAME=0.0.0.0

# Go API
COPY --from=backend-build /bin/server /app/server

# Next.js standalone server
COPY --from=frontend-build /src/.next/standalone /app/web
COPY --from=frontend-build /src/.next/static /app/web/.next/static
COPY --from=frontend-build /src/public /app/web/public

COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh \
    && addgroup -S app && adduser -S app -G app \
    && mkdir -p /data/uploads && chown -R app:app /data /app

USER app
EXPOSE 3000
CMD ["/app/start.sh"]
