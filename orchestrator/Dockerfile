FROM golang:1.21

WORKDIR /app


COPY go.mod go.sum ./

RUN GOPROXY=https://goproxy.cn go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux  go build -o ./orchestrator ./cmd/orchestrator/main.go

CMD ["./orchestrator"]
