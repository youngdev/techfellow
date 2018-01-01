FROM golang:alpine3.7

RUN apk update
RUN apk add git curl

RUN go get -u github.com/gocolly/colly/...

RUN mkdir -p /opt/cmcworkplace
COPY ./coinmarketcap.go /opt/cmcworkplace/

WORKDIR /opt/cmcworkplace/

RUN go build ./coinmarketcap.go

ENTRYPOINT ["./coinmarketcap"]
