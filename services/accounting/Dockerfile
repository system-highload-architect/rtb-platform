FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/accounting
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /accounting ./services/accounting/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /accounting .
COPY services/accounting/configs/ ./configs/
EXPOSE 9002
ENTRYPOINT ["./accounting"]