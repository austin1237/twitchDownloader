FROM golang:latest
RUN apt-get update
RUN apt-get install -y python-pip
RUN pip install -U livestreamer
ADD . /go/src/github.com/austin1237/twitchDownloader
WORKDIR /go/src/github.com/austin1237/twitchDownloader
RUN go build
ENTRYPOINT ["./twitchDownloader"]
CMD "/bin/bash"

