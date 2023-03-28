FROM golang:1.18-alpine
RUN apk update && apk add git

RUN mkdir /go/app
WORKDIR /go/app
ADD . /go/app