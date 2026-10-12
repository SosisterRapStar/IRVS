FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO выключен -> статический бинарник, который запустится в пустом образе.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app .

FROM alpine:3.22

RUN adduser -D -u 10001 app

COPY --from=builder /out/app /usr/local/bin/app

USER app
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD wget -q -O - http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/app"]
