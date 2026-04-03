FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=direct
ENV GOSUMDB=off

RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o server ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add nginx openssl supervisor

RUN mkdir -p /app/data /var/log/supervisor /etc/supervisor.d /etc/ssl/private /etc/ssl/certs

COPY certs/server.crt /etc/ssl/certs/server.crt
COPY certs/server.key /etc/ssl/private/server.key

COPY --from=builder /app/server /app/server
COPY --from=builder /go/bin/goose /app/goose
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/docs /app/docs

COPY nginx/nginx.conf /etc/nginx/http.d/default.conf
COPY supervisord.conf /etc/supervisord.conf

COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 8888

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
