FROM golang:1.16-alpine AS build
RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init

WORKDIR /go/src/iitk-coin

COPY . .

RUN go mod download

RUN go build
EXPOSE 8080

CMD ["./iitk-coin"]

