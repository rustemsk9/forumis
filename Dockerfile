FROM golang:1.13.6

RUN mkdir -p $GOPATH/src/forum
WORKDIR $GOPATH/src/forum
COPY . $GOPATH/src/forum
# RUN go get github.com/mattn/sqlite3
RUN go build

EXPOSE 8080

ENTRYPOINT ["./forum"]

