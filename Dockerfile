FROM golang:1.22-alpine AS build-go
ENV GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM node:22-alpine AS build-web
WORKDIR /src/admin
COPY web/admin/package.json web/admin/package-lock.json* ./
RUN npm install --registry=https://registry.npmmirror.com
COPY web/admin/ ./
RUN npm run build && mkdir -p /out/admin && cp -r dist/. /out/admin/
WORKDIR /src/mobile
COPY web/mobile/package.json web/mobile/package-lock.json* ./
RUN npm install --registry=https://registry.npmmirror.com
COPY web/mobile/ ./
RUN npm run build && mkdir -p /out/mobile && cp -r dist/. /out/mobile/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /app/data
ENV TZ=Asia/Shanghai \
    PORT=8080 \
    DEVICE_PORT=8768 \
    DB_PATH=/app/data/charginghub.db \
    STATIC_DIR=/app/web/admin \
    MOBILE_DIR=/app/web/mobile
COPY --from=build-go /out/server /app/server
COPY --from=build-web /out/admin /app/web/admin
COPY --from=build-web /out/mobile /app/web/mobile
EXPOSE 8080 8768
VOLUME ["/app/data"]
WORKDIR /app
CMD ["/app/server"]
