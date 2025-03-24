FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o larets .

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    git \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 get_helm.sh \
    && ./get_helm.sh \
    && rm get_helm.sh

WORKDIR /app

COPY --from=builder /app/larets .

RUN mkdir -p /app/storage/docker /app/storage/git /app/storage/helm /app/storage/temp

COPY .env* .env

EXPOSE 8080

VOLUME ["/app/storage"]

CMD ["./larets"]