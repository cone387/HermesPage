FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o hermespage .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/hermespage .
COPY web/ ./web/
RUN mkdir -p /app/reports
VOLUME /app/reports
EXPOSE 8080
CMD ["./hermespage", "serve"]
