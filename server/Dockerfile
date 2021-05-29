FROM golang:1.16

WORKDIR /go/src/app
COPY . .

RUN go get ./...
RUN go build .

CMD ["/go/src/app/main"]
