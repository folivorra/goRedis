# контейнер для сборки с компилятором
FROM golang:1.24.3-alpine AS builder

# установка git`a
RUN apk add --no-cache git

# рабочая директория
WORKDIR /app

# копируем go.mod go.sum и устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# копируем весь код
COPY . .

# компилируем бинарник
RUN go build -o app ./cmd

# чистый контейнер без мусора и компилятора
FROM alpine:latest

# новая рабочая директория
WORKDIR /root

# копируем из билдера бинарник
COPY --from=builder /app/app .

# открываем порт
EXPOSE 8080

# устанавливаем команду запуска
CMD ["./app"]
