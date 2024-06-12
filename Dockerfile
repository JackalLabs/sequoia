FROM golang:1.21

RUN apt update
RUN apt install yq -y

ADD . sequoia

WORKDIR sequoia

# RUN make install

COPY go.mod go.sum ./
RUN go mod download
RUN go get

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sequoia


# RUN ./scripts/sequoia.sh


CMD ["testrun""]
