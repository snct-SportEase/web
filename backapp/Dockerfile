# builder stage
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# Route public Go module metadata through Takumi Guard during image builds.
ENV GOPROXY=https://golang.flatt.tech

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/main.go

# final stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache mysql-client \
    && addgroup -S app \
    && adduser -S -D -H -G app app \
    && mkdir -p /app/uploads \
    && chown -R app:app /app

COPY --from=builder --chown=app:app /app/server ./server

EXPOSE 8080

USER app

CMD ["./server"]
