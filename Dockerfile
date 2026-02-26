FROM golang:1.24.1-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o raft-node ./src/main

CMD ["./raft-node"]