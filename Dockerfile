FROM golang:1.25-alpine AS builder
WORKDIR /usr/src/app
COPY . .
RUN go build -o main cmd/main.go

FROM golang:1.25-alpine
WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/config.yaml /usr/src/app/config.yaml
COPY --from=builder /usr/src/app/main /usr/src/app/main

CMD ["./main"]
