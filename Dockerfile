FROM golang:1.9


WORKDIR /go/src/github.com/jberkhahn/kyalia

COPY . .

RUN go install -v ./...

CMD ["kyalia"]
