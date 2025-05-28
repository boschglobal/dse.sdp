FROM golang:bullseye AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /html2img main.go

FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y wget gnupg ca-certificates && \
    wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && \
    apt install -y ./google-chrome-stable_current_amd64.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /html2img /usr/local/bin/html2img
WORKDIR /work
ENTRYPOINT ["html2img"]
