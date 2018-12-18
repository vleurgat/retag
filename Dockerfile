FROM golang:alpine AS build
RUN apk add git
ENV GOPATH=/go
ADD . /go/src/github.com/vleurgat/retag
WORKDIR /go/src/github.com/vleurgat/retag
RUN go get -d ./...
RUN go install github.com/vleurgat/retag/cmd/retag

FROM alpine
WORKDIR /usr/local/bin
COPY --from=build /go/bin/retag /usr/local/bin
ENTRYPOINT ["/usr/local/bin/retag"]
