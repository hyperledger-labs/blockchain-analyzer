FROM golang:1.12.7

# Python 2.7
RUN apt update && apt-get install python -y
RUN apt install virtualenv -y

# Pre-built executable
COPY ./agent/fabricbeat/prebuilt/Ubuntu-bionic-18.04 $GOPATH/src/github.com/blockchain-analyzer/agent/fabricbeat

WORKDIR $GOPATH/src/github.com/blockchain-analyzer/agent/fabricbeat

ENTRYPOINT chown root fabricbeat.yml && ./fabricbeat -e -d "*"

