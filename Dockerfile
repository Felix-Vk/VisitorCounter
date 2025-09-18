FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod tidy
COPY . .
RUN go build -o visitor-counter main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/visitor-counter .
COPY --from=builder /app/stats.json .
EXPOSE 8080
CMD ["./visitor-counter"]
