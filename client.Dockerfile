FROM golang:1.18

WORKDIR /file-transfer

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal

RUN rm -rf ./cmd/server
RUN rm -rf ./internal/server
RUN mkdir -p transfer-files

RUN go mod download
RUN go build ./cmd/client
ENTRYPOINT ./client -address $SERVER_ADDRESS -port $SERVER_PORT -file /file-transfer/transfer-files/$FILE_NAME