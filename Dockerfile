FROM golang:1.20-bullseye

WORKDIR /go/src
COPY ./ .

RUN go mod download
RUN go install github.com/cosmtrek/air@latest

CMD ["air", "-c", ".air.toml"]
