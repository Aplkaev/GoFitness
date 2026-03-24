FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
ENV GODEBUG netdns=go

RUN go mod download

COPY . .


RUN go build -o main ./bot

EXPOSE 8080

CMD ["./main"]