FROM golang

COPY . .

RUN go mod download

CMD ["go run main.go"]