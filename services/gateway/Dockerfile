# Стадия 1: сборка дашборда (Node.js)
FROM node:22-alpine AS dashboard-builder
WORKDIR /app
COPY services/dashboard/package.json services/dashboard/package-lock.json ./
RUN npm install
COPY services/dashboard/ .
RUN npm run build

# Стадия 2: сборка Gateway (Go)
FROM golang:1.25-alpine AS gateway-builder
WORKDIR /app
COPY . .
WORKDIR /app/services/gateway
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./services/gateway/cmd

# Финальный образ
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=gateway-builder /gateway .
COPY --from=dashboard-builder /app/dist ./web/dist
COPY services/gateway/configs/ ./configs/
EXPOSE 8080
ENTRYPOINT ["./gateway"]