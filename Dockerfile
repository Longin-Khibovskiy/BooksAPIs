FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates postgresql-client tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder /app/main .
COPY templates ./templates
COPY stylesheets ./stylesheets
COPY migrations ./migrations
COPY entrypoint.sh .

RUN chmod +x entrypoint.sh && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./main"]
