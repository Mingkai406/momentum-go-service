FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY *.go ./

RUN go mod download
RUN go build -o scheduler .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/scheduler .

EXPOSE 8080

CMD ["./scheduler"]
