# 1)собираем Go приложения
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# копируем go.mod и gu.sum
COPY go.mod go.sum ./
RUN go mod download

# копируем исходный код
 COPY . .

# Собираем приложение 
RUN go build -o server ./cmd/main.go

# 2)финальный образ
FROM alpine:latest

# устанавливаем Nginx, SQLite, OpenSSL
RUN apk --no-cache add nginx sqlite openssl supervisor

# Создаём SSL сертификаты для HTTPS
RUN openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/ssl/private/server.key \
    -out /etc/ssl/certs/server.crt \
    -subj "C=RU/ST=Moscow/L=Moscow/O=PaymentSystem/CN=localhost"

# копируем бинарник из этапа сборки
COPY --from=builder /app/migrations /app/migrations

# копируем конфиг Nginx 
COPY nginx/nginx.conf /etc/nginx/http.d/default.conf

# создаем папку для базы данных
 RUN mkdir -p /app/data

 #открываем порыт
 EXPOSE 8888

 # скрипт запуска  (ДОДЕЛАТЬ)
 CMD ["/app/server"]