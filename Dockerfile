FROM golang:1.25.7-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
COPY config ./config

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -tags timetzdata -ldflags="-s -w" -o /out/gorillo ./cmd/gorillo
RUN mkdir -p /out/data

FROM scratch
WORKDIR /app

COPY --from=builder /out/gorillo /app/gorillo
COPY --from=builder /src/web /app/web
COPY --from=builder /src/config/config.docker.toml /app/config/config.toml
COPY --from=builder /out/data /app/data

EXPOSE 8000

ENTRYPOINT ["/app/gorillo"]
