FROM ubuntu:18.10
MAINTAINER wcosta@mozilla.com

ENV GOPATH /home/go
ENV PATH $PATH:$GOPATH/bin

RUN apt-get update
RUN apt-get install -y jq curl golang gcc git

RUN go get github.com/herloct/url-parser
RUN go get github.com/taskcluster/taskcluster-proxy
RUN go get github.com/taskcluster/livelog

ADD entrypoint.sh /usr/bin/

ENTRYPOINT ["entrypoint.sh"]
