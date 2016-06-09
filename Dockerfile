FROM golang:latest

WORKDIR /go/src/github.com/warmans/stressfaktor-api
ADD . .

RUN make build

ENTRYPOINT ["stressfaktor-api"]

EXPOSE 8080
