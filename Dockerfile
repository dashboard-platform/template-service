FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./template ./cmd/main.go

EXPOSE 8080

RUN chmod +x template

CMD ["./template"]