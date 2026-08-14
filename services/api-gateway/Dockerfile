FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /app/main .
USER 10001
EXPOSE 8080
CMD ["./main"]
