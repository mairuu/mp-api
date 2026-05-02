FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev libwebp-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o api ./cmd/api
RUN CGO_ENABLED=1 GOOS=linux go build -o migrate ./cmd/migrate

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrate .

EXPOSE 8080

CMD ["./api"]