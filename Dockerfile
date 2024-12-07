FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o click-telegram-mini-server

CMD ["/app/click-telegram-mini-server"]