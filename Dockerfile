# syntax=docker/dockerfile:1
FROM golang:1.21

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o scheduler ./cmd/scheduler

CMD ["./scheduler"]
