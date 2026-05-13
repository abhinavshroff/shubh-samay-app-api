# Multi-stage build for tiny final image
FROM golang:1.23-alpine AS builder
WORKDIR /build
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/api .
COPY --from=builder /build/ephemeris ./ephemeris
ENV TZ=Asia/Kolkata
EXPOSE 8080
CMD ["./api"]
