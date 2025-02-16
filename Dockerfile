FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o avito-shop ./cmd/main.go

RUN chmod +x ./avito-shop

EXPOSE 8080

CMD ["./avito-shop"]