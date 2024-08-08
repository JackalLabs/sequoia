#!/bin/bash

set -eu

source ./scripts/setup-chain.sh

UPGRADE_HEIGHT="60"

OLD_CHAIN_VER="v4.0.3"
NEW_CHAIN_VER="marston/recovery"

sender="jkl10k05lmc88q5ft3lm00q30qkd9x6654h3lejnct"
current_dir=$(pwd)
TMP_ROOT="$(dirname $(pwd))/_build"
mkdir -p "${TMP_ROOT}"


bypass_go_version_check () {
    VER=$(go version)
    sed -i -e 's/! go version | grep -q "go1.2[0-9]"/false/' ${1}/Makefile
}

install_old () {
    TMP_GOCACHE="${TMP_ROOT}"/gocache
    mkdir -p "${TMP_GOCACHE}"

    echo "installing v4 canined"
    if [[ ! -e "${TMP_ROOT}/canine-chain" ]]; then
        cd ${TMP_ROOT}
        git clone https://github.com/JackalLabs/canine-chain.git
        cd ${current_dir}
    fi

    PROJ_DIR="${TMP_ROOT}/canine-chain"


    cd ${PROJ_DIR}
    git fetch
    git switch tags/${OLD_CHAIN_VER} --detach
    bypass_go_version_check ${PROJ_DIR}
    make install
    git restore Makefile
    echo "finished chain installation"
    
    cd "${current_dir}"
    canined version

    
    if [[ ! -e "${TMP_ROOT}/sequoia" ]]; then
        cd ${TMP_ROOT}
        git clone https://github.com/JackalLabs/sequoia.git
        cd sequoia
        cd ${current_dir}
    fi
}

install_new_chain () {
    PROJ_DIR="${TMP_ROOT}/canine-chain"
    git fetch
    cd ${PROJ_DIR}
    #git switch tags/${NEW_CHAIN_VER} --detach
    git checkout ${NEW_CHAIN_VER}

    bypass_go_version_check ${PROJ_DIR}
    make install
    git restore Makefile


    cd "${current_dir}"
    canined version
}

start_chain () {
    startup
    from_scratch
    fix_config

    screen -L -Logfile chain_log.log  -d -m -S "canined" bash -c "canined start --pruning=nothing --minimum-gas-prices=0ujkl"
}

set_upgrade_prop () {
    canined tx gov submit-proposal software-upgrade "v410" --upgrade-height ${UPGRADE_HEIGHT} --upgrade-info "tmp" --title "v4 Upgrade" \
        --description "upgrade" --from charlie -y --deposit "20000000ujkl"

    sleep 6

    canined tx gov vote 1 yes --from ${KEY} -y
    echo "voting successful"
}

upgrade_chain () {
    while true; do
        BLOCK_HEIGHT=$(canined status | jq '.SyncInfo.latest_block_height' -r)
        if [ $BLOCK_HEIGHT = "$UPGRADE_HEIGHT" ]; then
            # assuming running only 1 canined
            echo "BLOCK HEIGHT = $UPGRADE_HEIGHT REACHED, KILLING OLD ONE"
            killall canined
            break
        else
            canined q gov proposal 1 --output=json
            echo "BLOCK_HEIGHT = $BLOCK_HEIGHT"
            sleep 2
        fi
    done

    install_new_chain
}

restart_chain () {
    screen -L -Logfile chain_log.log -d -m -S "canined" bash -c "canined start --pruning=nothing --minimum-gas-prices=0ujkl"
}


init_sequoia () {
    rm -rf $HOME/providers/sequoia${1}
    sequoia init --home="$HOME/providers/sequoia${1}"

#    sed -i -e 's/rpc_addr: https:\/\/jackal-testnet-rpc.polkachu.com:443/rpc_addr: tcp:\/\/localhost:26657/g' $HOME/providers/sequoia${1}/config.yaml
#    sed -i -e 's/grpc_addr: jackal-testnet-grpc.polkachu.com:17590/grpc_addr: localhost:9090/g' $HOME/providers/sequoia${1}/config.yaml

    sed -i -e 's/data_directory: $HOME\/.sequoia\/data/data_directory: $HOME\/providers\/sequoia0\/data/g' $HOME/providers/sequoia${1}/config.yaml
}


start_sequoia () {
  pwd
  cp -r ./scripts/storage "$HOME/providers/sequoia${1}/storage"
  screen -L -Logfile sequoia.log  -d -m -S "sequoia" bash -c "sequoia jprovd-salvage $HOME/providers/sequoia${1} --home=$HOME/providers/sequoia${1}"
}

recycle () {
    sequoia recycle
}

shut_down () {
    killall canined sequoia
}

install_old

start_chain
echo "CHAIN STARTED!!!"
sleep 5

start_sequoia 0
echo "provider started!!!"
sleep 35

set_upgrade_prop
sleep 5
canined q gov proposal 1

echo "upgrading chain to v410"
upgrade_chain
restart_chain
sleep 5


read -rsp $'Press any key to shutdown...\n' -n1 key

shut_down
#cleanup
