FROM golang:1.21

ADD . sequoia

WORKDIR sequoia

RUN make install

RUN ./scripts/sequoia.sh


CMD ["testrun""]
