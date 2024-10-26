FROM golang:1.19

LABEL authors="viste"

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o main .

CMD ["./main"]