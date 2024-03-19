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

In order for your provider to run correctly, you will need to set up a domain for your provider pointed at the port your provider is running on and set that up in the `config.yaml`. You will also need to make sure you have port `4001` open on TCP and UDP for IPFS support, or you could be penalized by the reporting system.