FROM golang:1.26-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

FROM alpine:3.21
RUN adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=build /bin/server /app/server
RUN mkdir -p /app/uploads && chown appuser /app/uploads
USER appuser
EXPOSE 8090
ENTRYPOINT ["/app/server"]
