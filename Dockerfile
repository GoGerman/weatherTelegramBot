FROM golang

COPY . /weatherTelegramBot

WORKDIR /weatherTelegramBot

RUN go get -d -v ./...

RUN go install -v ./...

CMD ["bash", "-c", "go run main.go"]