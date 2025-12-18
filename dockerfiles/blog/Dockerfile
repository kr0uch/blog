FROM golang:1.25.1-alpine AS builder

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/blog ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/blog /bin/blog
COPY --from=builder /app/migrations/ /migrations/
COPY --from=builder /app/.env /.env

CMD ["/bin/blog"]
