FROM golang:1.24

ARG RPC_ADDR="http://localhost:26657"
ARG GRPC_ADDR="localhost:9090"

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make install

# RPC GRPC api ipfs
EXPOSE 26657 9090 3333 4005

WORKDIR /root/.sequoia

RUN sequoia init

RUN sed -i -e "s/http:\/\/localhost:26657/${RPC_ADDR}/" config.yaml
RUN sed -i -e "s/localhost:9090/${GRPC_ADDR}/" config.yaml

CMD ["sequoia", "start"]
