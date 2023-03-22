FROM golang:1.16-alpine

WORKDIR /go/src/app

COPY . .

RUN apk add --no-cache git
RUN go get -d -v ./...

RUN go build -o app

EXPOSE 8080

CMD ["./app"]