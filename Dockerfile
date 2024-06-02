FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/main .

FROM alpine:3.20

COPY --from=builder /app/main /app/main
RUN mkdir "/etc/ichat"

WORKDIR /app

CMD ["./main"]
