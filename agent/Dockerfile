FROM golang:1.21

WORKDIR /app


COPY go.mod go.sum ./

RUN  GOPROXY=https://goproxy.cn go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn go build -o ./agent ./cmd/agent/main.go

CMD ["./agent"]
