# syntax=docker/dockerfile:1

# Start from a base Golang image to build the binary
FROM golang:1.23 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /lib

COPY ./setup-chain.sh .

RUN wget https://github.com/CosmWasm/wasmvm/raw/v1.2.6/internal/api/libwasmvm.x86_64.so

WORKDIR /app

RUN git clone https://github.com/JackalLabs/canine-chain.git .

RUN git checkout $(git describe --abbrev=0 --tags)

RUN make build

# Use a smaller image to run the binary
FROM debian:bookworm-slim

WORKDIR /root

RUN apt-get update && apt-get install -y \
	build-essential \
	git \
	curl \
	bash \
	wget \
	jq

COPY --from=builder /app/build/canined /usr/local/bin/canined
COPY --from=builder /lib/libwasmvm.x86_64.so /lib/libwasmvm.x86_64.so
COPY --from=builder /lib/setup-chain.sh /root/setup-chain.sh

WORKDIR /root

RUN ./setup-chain.sh

# Expose default RPC, P2P, and REST ports
EXPOSE 26656 26657 1317 9092 9090

WORKDIR /root/.canine

CMD ["canined", "start"]
