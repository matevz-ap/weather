FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o /app/cmd/weather /app/cmd/main.go

EXPOSE 8080

CMD ["/app/cmd/weather"]