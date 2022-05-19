FROM golang:1.17-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd/
COPY internal ./internal/

RUN go build -o ./cibus-coupons-telegram-bot ./cmd

CMD [ "./cibus-coupons-telegram-bot" ]
