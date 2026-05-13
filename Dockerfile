# Multi-stage build for the Go API.
# The panchang calculator uses github.com/mshafiee/swephgo, which links
# against the native Swiss Ephemeris library (-lswe). Debian packages that
# library; Alpine does not provide it by default, which causes linker errors.
FROM golang:1.23-bookworm AS builder
WORKDIR /build
RUN apt-get update \
    && apt-get install -y --no-install-recommends gcc libc6-dev libswe-dev \
    && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# libswe uses libm symbols (sin, cos, sqrt, etc.); pass -lm through cgo so
# the external linker resolves them after -lswe during the final link step.
RUN CGO_LDFLAGS="-lm" CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata libswe2.0 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /build/api .
COPY --from=builder /build/ephemeris ./ephemeris
ENV TZ=Asia/Kolkata
EXPOSE 8080
CMD ["./api"]
