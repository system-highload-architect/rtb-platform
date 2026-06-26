FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/analytics
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics ./services/analytics/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /analytics .
COPY services/analytics/configs/ ./configs/
EXPOSE 9003
ENTRYPOINT ["./analytics"]