FROM golang:1.23-bookworm AS builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /promstock .

FROM debian:bookworm-slim AS run
WORKDIR /app

RUN apt-get update -y && apt-get install ca-certificates -y

COPY --from=builder /promstock ./promstock

CMD ["./promstock"]