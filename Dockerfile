FROM golang

ADD . /go/src/github.com/eluleci/lightning

RUN go get github.com/eluleci/lightning
RUN go install github.com/eluleci/lightning

ENTRYPOINT /go/bin/lightning

EXPOSE 8080

# HOW TO RUN IMAGE
# docker build -t thunderdock .
# docker run --publish 6060:8080 --name thunderdock -rm thunderdock