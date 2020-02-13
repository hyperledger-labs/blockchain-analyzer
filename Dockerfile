FROM golang:1.13.0

COPY ./.git $GOPATH/src/github.com/blockchain-analyzer/.git
COPY ./agent $GOPATH/src/github.com/blockchain-analyzer/agent

WORKDIR $GOPATH/src/github.com/blockchain-analyzer/agent/fabricbeat

RUN rm go.mod go.sum

RUN apt-get update && apt-get install make -y
RUN make go-get
RUN apt-get update && apt-get install virtualenv -y
RUN make update
RUN make

ENTRYPOINT chown root fabricbeat.yml && chmod 644 fabricbeat.yml && ./fabricbeat -e -d "*"

