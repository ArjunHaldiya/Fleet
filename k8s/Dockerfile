from golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build - o fleet .

FROM alpine:latest
RUN apk --no-cache add ca-certificate
WORKDIR /root/
COPY --from=builder /app/fleet .
EXPOSE 8080
CMD ["./fleet"]