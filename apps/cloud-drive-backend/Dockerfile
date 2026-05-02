FROM golang:1.25-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,id=go-mod-cache go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build,id=go-build-cache CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go

FROM alpine:3.22

WORKDIR /app
RUN apk add --no-cache ca-certificates wget \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /app/server

USER appuser

EXPOSE 9000

CMD ["/app/server"]
