FROM golang:1.18

WORKDIR /file-transfer

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal

RUN rm -rf ./cmd/client
RUN rm -rf ./internal/client
RUN mkdir -p uploads

EXPOSE $SERVER_PORT

RUN go mod download
RUN go build ./cmd/server
ENTRYPOINT ./server -port $SERVER_PORT