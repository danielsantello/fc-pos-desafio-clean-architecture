FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -o /out/ordersystem \
    ./cmd/ordersystem

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/ordersystem /app/ordersystem

USER 65532:65532

EXPOSE 8000 50051 8080

ENTRYPOINT ["/app/ordersystem"]
