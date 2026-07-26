FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mimir ./cmd/mimir

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D mimir
COPY --from=builder /mimir /usr/local/bin/mimir
USER mimir
EXPOSE 8420
ENV MIMIR_CORTEX_BACKEND=surreal
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8420/health || exit 1
ENTRYPOINT ["mimir", "serve", "--addr=:8420"]
