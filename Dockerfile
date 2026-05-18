FROM golang:1.26-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
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
EXPOSE 5487
ENV HERMES_JWT_SECRET=change-me-in-production
CMD ["./hermespage", "serve"]
