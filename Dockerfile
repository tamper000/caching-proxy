# Stage 1
FROM golang:1.24.4-bookworm as builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o main cmd/main.go

# Stage 2
FROM alpine

WORKDIR /app

COPY --from=builder /src/main .

EXPOSE 8080

ENTRYPOINT ["./main"]