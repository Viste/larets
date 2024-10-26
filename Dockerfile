FROM golang:1.22

LABEL authors="viste"

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o main ./cmd/server

CMD ["./main"]