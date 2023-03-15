FROM golang

COPY . .

RUN go mod download

RUN go build -o main .

CMD ["./main"]