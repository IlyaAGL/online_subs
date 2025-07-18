FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o sub_migrator ./cmd/migrator

CMD ["./sub_migrator"]