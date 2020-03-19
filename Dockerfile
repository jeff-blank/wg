FROM golang:1.13.8 as builder

RUN apt-get update && apt-get -y upgrade
RUN go get golang.org/dl/go1.12.17 && /go/bin/go1.12.17 download && /go/bin/go1.12.17 get honnef.co/go/js/dom
RUN go get github.com/revel/revel \
	github.com/revel/cmd/revel \
	github.com/gopherjs/gopherjs \
	github.com/gopherjs/jquery \
	github.com/wcharczuk/go-chart

COPY app /go/src/github.com/jeff-blank/wg/app/
COPY conf /go/src/github.com/jeff-blank/wg/conf/
COPY jq /go/src/github.com/jeff-blank/wg/jq/
COPY public /go/src/github.com/jeff-blank/wg/public/
COPY Makefile /go/src/github.com/jeff-blank/wg/

WORKDIR /go/src/github.com/jeff-blank/wg

ARG NR_LICENSE_B
ARG REVEL_SECRET_B
ARG DB_CONNECT_PROD_B

ENV NR_LICENSE=$NR_LICENSE_B REVEL_SECRET=$REVEL_SECRET_B DB_CONNECT_PROD=$DB_CONNECT_PROD_B

RUN sed -i 's@gmake@GOPHERJS_GOROOT=`/go/bin/go1.12.17 env GOROOT` make@' Makefile
RUN NR_LICENSE="$NR_LICENSE_B" \
	REVEL_SECRET=$REVEL_SECRET_B \
	DB_CONNECT_PROD="$DB_CONNECT_PROD" \
	CGO_ENABLED=0 GOOS=linux \
	make release RELEASEDEST=/wg REINPLACE='sed -i'

FROM debian:buster-slim

COPY --from=builder /wg /wg

ENTRYPOINT ["/wg/run.sh"]
