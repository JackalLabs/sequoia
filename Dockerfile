FROM golang:1.21

RUN apt update
RUN apt install yq -y

ADD . sequoia

WORKDIR sequoia

RUN make install

RUN ./scripts/sequoia.sh


CMD ["testrun""]
