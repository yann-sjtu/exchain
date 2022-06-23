#!/bin/bash

KEY="captain"
CHAINID="exchain-67"
MONIKER="okc"
CURDIR=`dirname $0`
HOME_SERVER=$CURDIR/"_cache_evm"

set -e
set -o errexit
set -a
set -m


killbyname() {
  NAME=$1
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2", "$8}'
  ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2}' | sh
  echo "All <$NAME> killed!"
}


run() {
    LOG_LEVEL=main:debug,iavl:info,*:error,state:info,provider:info

    exchaind start --pruning=nothing --rpc.unsafe \
      --local-rpc-port 26657 \
      --log_level $LOG_LEVEL \
      --log_file json \
      --enable-dynamic-gp=false \
      --mempool.sort_tx_by_gp=true \
      --consensus.timeout_commit 50ms \
      --disable-abci-query-mutex=true \
      --mempool.max_tx_num_per_block=10000 \
      --mempool.size=100000 \
      --local_perf=tx \
      --enable-preruntx=false \
      --iavl-enable-async-commit \
      --enable-gid \
      --append-pid=true \
      --iavl-commit-interval-height 10 \
      --iavl-output-modules evm=0,acc=0 \
      --trace --home $HOME_SERVER --chain-id $CHAINID \
      --elapsed Round=1,CommitRound=1,Produce=1 \
      --rest.laddr "tcp://localhost:8545" > okc.txt 2>&1 &

# --iavl-commit-interval-height \
# --iavl-enable-async-commit \
#      --iavl-cache-size int                              Max size of iavl cache (default 1000000)
#      --iavl-commit-interval-height int                  Max interval to commit node cache into leveldb (default 100)
#      --iavl-debug int                                   Enable iavl project debug
#      --iavl-enable-async-commit                         Enable async commit
#      --iavl-enable-pruning-history-state                Enable pruning history state
#      --iavl-height-orphans-cache-size int               Max orphan version to cache in memory (default 8)
#      --iavl-max-committed-height-num int                Max committed version to cache in memory (default 8)
#      --iavl-min-commit-item-count int                   Min nodes num to triggle node cache commit (default 500000)
#      --iavl-output-modules
    exit
}


killbyname exchaind
killbyname exchaincli

set -x # activate debugging

# run

# remove existing daemon and client
rm -rf ~/.exchain*
rm -rf $HOME_SERVER

(cd .. && make install VenusHeight=1)

# Set up config for CLI
exchaincli config chain-id $CHAINID
exchaincli config output json
exchaincli config indent true
exchaincli config trust-node true
exchaincli config keyring-backend test

# if $KEY exists it should be deleted
#
#    "eth_address": "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F",
#     prikey: 8ff3ca2d9985c3a52b459e2f6e7822b23e1af845961e22128d5f372fb9aa5f17
exchaincli keys add --recover captain -m "puzzle glide follow cruel say burst deliver wild tragic galaxy lumber offer" -y

#    "eth_address": "0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0",
exchaincli keys add --recover admin16 -m "palace cube bitter light woman side pave cereal donor bronze twice work" -y

exchaincli keys add --recover admin17 -m "antique onion adult slot sad dizzy sure among cement demise submit scare" -y

exchaincli keys add --recover admin18 -m "lazy cause kite fence gravity regret visa fuel tone clerk motor rent" -y

# 0xEb4D58BA0ADBD52AE9dB5a97E5c99289e721b30C
# afbf867c175d639c46199954f8f3c126f00a45c232089d9ce72b3555ebf6d02d
exchaincli keys add --recover admin19 -m "solid dinner float window exile lend comfort thunder tourist argue retreat edge" -y

# 0xAdD3292c9479f0D3faffB50490adF3BF90361A74
# b97487ce7527f6f638b475936edacda3e7365f394d3a453a9e16a8c71de126fd
exchaincli keys add --recover admin20 -m "asthma sea appear divorce someone awkward behave swamp candy record squeeze north" -y

# 0x37f3BbE03880B3062B24F4a4b6c17F1188a1c8E5
# 424359a24b08bb6f1bd63851299be67c813dc2cf295e42d27b8bcf7a887af8dd
exchaincli keys add --recover admin21 -m "midnight mistake there culture idea speed brisk tunnel legend sniff excuse anger" -y

# 0xFE157eB5bf53bdF0CA11C4b8D7467a022b195535
# a1fa94c958dfaa60f395ad334bdce3997269e828e5b4b90cecf94b3e01b1647b
exchaincli keys add --recover admin22 -m "dove tobacco token resemble near rate please sheriff priority sure twin tag" -y


#exchaincli keys add --recover admin19 -m "" -y
# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
exchaind init $MONIKER --chain-id $CHAINID --home $HOME_SERVER

# Change parameter token denominations to okt
cat $HOME_SERVER/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json

# Enable EVM

if [ "$(uname -s)" == "Darwin" ]; then
    sed -i "" 's/"enable_call": false/"enable_call": true/' $HOME_SERVER/config/genesis.json
    sed -i "" 's/"enable_create": false/"enable_create": true/' $HOME_SERVER/config/genesis.json
    sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' $HOME_SERVER/config/genesis.json
else 
    sed -i 's/"enable_call": false/"enable_call": true/' $HOME_SERVER/config/genesis.json
    sed -i 's/"enable_create": false/"enable_create": true/' $HOME_SERVER/config/genesis.json
    sed -i 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' $HOME_SERVER/config/genesis.json
fi

# Allocate genesis accounts (cosmos formatted addresses)
exchaind add-genesis-account $(exchaincli keys show $KEY    -a) 100000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin16 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin17 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin18 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin19 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin20 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin21 -a) 900000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin22 -a) 900000000okt --home $HOME_SERVER


# Sign genesis transaction
exchaind gentx --name $KEY --keyring-backend test --home $HOME_SERVER

# Collect genesis tx
exchaind collect-gentxs --home $HOME_SERVER

# Run this to ensure everything worked and that the genesis file is setup correctly
exchaind validate-genesis --home $HOME_SERVER
exchaincli config keyring-backend test

run

# exchaincli tx send captain 0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0 1okt --fees 1okt -b block -y
