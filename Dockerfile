# Stage 1
FROM golang:1.24.4-bookworm as builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o main cmd/main.go

# Stage 2
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

RUN adduser -D -s /bin/sh appuser
COPY --from=builder /src/main .
RUN chown appuser:appuser main
USER appuser

EXPOSE 8080

ENTRYPOINT ["./main"]
