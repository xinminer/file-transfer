FROM golang:1.18

WORKDIR /file-transfer

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
RUN rm -rf ./cmd/server
COPY internal ./internal
RUN rm -rf ./internal/server

RUN mkdir -p transfer-files

CMD go run ./cmd/client -address $SERVER_ADDRESS -port $SERVER_PORT -file /transfer-files/$FILE_NAME