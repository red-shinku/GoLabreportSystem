# 用于构建
FROM golang:1.26.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

# 用于运行，基于alpine
FROM alpine:3.19
# ca-certificates 用于 HTTPS 客户端调用
# 包含了受信任 CA 列表
RUN apk add --no-cache ca-certificates
WORKDIR /app

# 仅复制二进制和默认配置
COPY --from=builder /app/server /app/server
COPY config.json /app/config.json
COPY --from=builder /app/html /app/html
COPY --from=builder /app/scripts /app/scripts

# 容器内部可用 8080 8443 端口

EXPOSE 8080 8443
ENTR

YPOINT ["/app/server"]
