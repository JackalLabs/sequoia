![logo](./.assets/logo.png)

# Sequoia

Sequoia is a light-weight, fast, and complete implementation of a [Jackal Protocol](https://github.com/JackalLabs/canine-chain) Storage Provider made for SSD-based storage and IPFS compatibility.

## Installation

### Binaries
Head over to the [Releases](https://github.com/JackalLabs/sequoia/releases) and download the binary for your operating system.

### Building from Source
```shell
git clone https://github.com/JackalLabs/sequoia.git
cd sequoia
go install
```

## Running Sequoia
Once installed, running `sequoia` will show you a list of available commands.

### Initializing the Provider
Running `sequoia init` for the first time will generate your configuration files, by default, at `~/.sequoia/` which will include `config.yaml` and `provider_wallet.json` for you to edit. If you wish to recover an existing proivider, copy the `provider_wallet.json` from the old provider to your new set-up.

After this, make sure you fund the wallet address (can be found by running `sequoia wallet address`) with over 10_000_000_000ujkl or 10k JKL tokens. The recommended amount is that plus 100 JKL.

### Starting
Once the wallet is funded, running `sequoia start` again will start the provider and go through the initialization process if it is a new machine. From here, keep an eye on the provider logs. Happy providing!

### Earning Rewards

In order for your provider to run correctly, you will need to set up a domain for your provider pointed at the port your provider is running on and set that up in the `config.yaml`. You will also need to make sure you have port `4005` (or whatever you specified in the config) open on TCP and UDP for IPFS support, or you could be penalized by the reporting system.

### Configuration

Default config looks like this:  
```yaml
######################
### Sequoia Config ###
######################

queue_interval: 10
proof_interval: 120
stray_manager:
    check_interval: 30
    refresh_interval: 120
    hands: 2
chain_config:
    bech32_prefix: jkl
    rpc_addr: http://localhost:26657
    grpc_addr: localhost:9090
    gas_price: 0.02ujkl
    gas_adjustment: 1.5
domain: https://example.com
total_bytes_offered: 1092616192
data_directory: $HOME/.sequoia/data
api_config:
    port: 3333
    ipfs_port: 4005
    ipfs_domain: dns4/ipfs.example.com/tcp/4001
proof_threads: 1000
data_store_config:
    directory: $HOME/.sequoia/datastore
    backend: flatfs

######################
```  
`data_directory`: file path | directory to store database (badger db) files  
#### `data_store_config`
sequoia uses ipfs go-datastore interface to store files  
`flatfs` stores raw block contents on disk  
`badgerds` a key value database that uses LSM tree to store and manage data  
currently only supports `flatfs` and `badgerds` data store backends  
`directory`: file path | directory for data store files  
`backend`: `flatfs` or `badgerds` | data store backend to store files  

> Using `badgerds` as backend requires data store directory to be same as `data_directory` as both uses badger db to store data.

