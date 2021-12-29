FROM golang:1.17.3 as builder

#RUN go get golang.org/dl/go1.12.17 && \
#	/go/bin/go1.12.17 download
#RUN GO111MODULE=on /go/bin/go1.12.17 get github.com/gopherjs/gopherjs@1.12.3+go1.12 \
#	github.com/gopherjs/jquery \
#	honnef.co/go/js/dom
#RUN go get github.com/revel/revel \
#	github.com/revel/cmd/revel \
#	github.com/lib/pq \
#	github.com/revel/modules/server-engine/newrelic \
#	github.com/wcharczuk/go-chart

ENV GO111MODULE=on

COPY go.mod go.sum /go/src/github.com/jeff-blank/wg/
COPY app /go/src/github.com/jeff-blank/wg/app/
COPY conf /go/src/github.com/jeff-blank/wg/conf/
COPY jq /go/src/github.com/jeff-blank/wg/jq/
COPY public /go/src/github.com/jeff-blank/wg/public/
COPY Makefile /go/src/github.com/jeff-blank/wg/
WORKDIR src/github.com/jeff-blank/wg

RUN go get github.com/gopherjs/jquery
RUN go install github.com/gopherjs/gopherjs@1.17.1+go1.17.3
RUN go install github.com/revel/cmd/revel


ARG NR_LICENSE_B
ARG REVEL_SECRET_B
ARG DB_CONNECT_PROD_B

ENV NR_LICENSE=$NR_LICENSE_B REVEL_SECRET=$REVEL_SECRET_B DB_CONNECT_PROD=$DB_CONNECT_PROD_B

RUN sed -i 's/gmake/make/' Makefile
RUN CGO_ENABLED=0 GOOS=linux \
	make release RELEASEDEST=/wg REINPLACE='sed -i'

FROM alpine:latest

COPY --from=builder /wg /wg
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ARG TZ_B
ENV TZ=$TZ_B

ENTRYPOINT ["/wg/run.sh"]
