
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .


FROM alpine:latest

WORKDIR /app


RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

COPY templates ./templates
COPY stylesheets ./stylesheets
COPY .env .env

EXPOSE 8000

CMD ["./main"]
