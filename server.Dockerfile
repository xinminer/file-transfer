FROM golang:1.18

WORKDIR /file-transfer

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
RUN rm -rf ./cmd/client
COPY internal ./internal
RUN rm -rf ./internal/client

RUN mkdir -p uploads

CMD go run ./cmd/server -port $SERVER_PORT