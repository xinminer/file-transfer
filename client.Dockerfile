FROM golang:1.18

WORKDIR /file-transfer

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY demo ./demo

ENV GOPROXY https://goproxy.cn,direct

RUN rm -rf ./cmd/server
RUN rm -rf ./internal/server
RUN mkdir -p transfer-files

RUN go mod download
RUN go build ./demo/main
ENTRYPOINT ./main