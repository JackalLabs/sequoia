FROM golang:1.21

RUN apt update
RUN apt install yq -y

ADD . sequoia

WORKDIR sequoia

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sequoia

CMD ["sh", "scripts/sequoia.sh"]
