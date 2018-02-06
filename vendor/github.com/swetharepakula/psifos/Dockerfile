FROM golang:1.9


WORKDIR /go/src/github.com/swetharepakula/psifos

COPY . .

RUN go install -v ./...

CMD ["psifos"]
