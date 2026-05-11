FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/sekai-game-api .


FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget && adduser -D -u 1000 app

WORKDIR /app
COPY --from=builder /out/sekai-game-api /app/sekai-game-api
COPY config.yaml /app/config.yaml

USER app
EXPOSE 11000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:11000/healthz || exit 1

ENTRYPOINT ["/app/sekai-game-api"]
CMD ["-config", "/app/config.yaml"]
