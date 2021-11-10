#!/bin/bash

KEY="captain"
CHAINID="exchainevm-8"
MONIKER="oec"
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
    LOG_LEVEL=main:info,*:error

    exchaind start --pruning=nothing  --mempool.enable_pending_pool=true --rpc.unsafe --paralleled-tx \
      --local-rpc-port 26657 \
      --consensus.timeout_commit 600ms \
      --iavl-enable-async-commit \
      --iavl-commit-interval-height 2 \
      --iavl-output-modules evm=1,acc=0 \
      --trace --home $HOME_SERVER --chain-id $CHAINID \
      --elapsed DeliverTxs=1 \
      --rest.laddr "tcp://localhost:8545" > oec.log 2>&1 &

    exit
}


killbyname exchaind
killbyname exchaincli

set -x # activate debugging


rm -rf ~/.exchain*
rm -rf $HOME_SERVER

(cd .. && make install)

exchaincli config chain-id $CHAINID
exchaincli config output json
exchaincli config indent true
exchaincli config trust-node true
exchaincli config keyring-backend test

#    "eth_address": "0xbbE4733d85bc2b90682147779DA49caB38C0aA1F",
exchaincli keys add --recover captain -m "puzzle glide follow cruel say burst deliver wild tragic galaxy lumber offer" -y

#    "eth_address": "0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0",
exchaincli keys add --recover admin16 -m "palace cube bitter light woman side pave cereal donor bronze twice work" -y

# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
exchaind init $MONIKER --chain-id $CHAINID --home $HOME_SERVER

# Change parameter token denominations to okt
cat $HOME_SERVER/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json
cat $HOME_SERVER/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="okt"' > $HOME_SERVER/config/tmp_genesis.json && mv $HOME_SERVER/config/tmp_genesis.json $HOME_SERVER/config/genesis.json

# Enable EVM
sed -i "" 's/"enable_call": false/"enable_call": true/' $HOME_SERVER/config/genesis.json
sed -i "" 's/"enable_create": false/"enable_create": true/' $HOME_SERVER/config/genesis.json



# scf


sed -i "" 's/create_empty_blocks = true/create_empty_blocks = false/' $HOME_SERVER/config/config.toml 
sed -i "" 's/size = 2000/size = 200000/' $HOME_SERVER/config/config.toml 
sed -i "" 's/max_tx_num_per_block = 300/max_tx_num_per_block = 3000/' $HOME_SERVER/config/config.toml 



sed -i "" 's/timeout_propose = "3s"/timeout_propose = "3s"/' $HOME_SERVER/config/config.toml 
sed -i "" 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "3s"/' $HOME_SERVER/config/config.toml 

sed -i "" 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/' $HOME_SERVER/config/config.toml 
sed -i "" 's/timeout_precommit = "1s"/timeout_precommit = "5s"/' $HOME_SERVER/config/config.toml 

sed -i "" 's/timeout_commit = "3s"/timeout_commit = "5s"/' $HOME_SERVER/config/config.toml 




# Allocate genesis accounts (cosmos formatted addresses)
exchaind add-genesis-account $(exchaincli keys show $KEY    -a) 100000000okt --home $HOME_SERVER
exchaind add-genesis-account $(exchaincli keys show admin16 -a) 900000000okt --home $HOME_SERVER

# Sign genesis transaction
exchaind gentx --name $KEY --keyring-backend test --home $HOME_SERVER

# Collect genesis tx
exchaind collect-gentxs --home $HOME_SERVER

# Run this to ensure everything worked and that the genesis file is setup correctly
exchaind validate-genesis --home $HOME_SERVER
exchaincli config keyring-backend test

run

# exchaincli tx send captain 0x83D83497431C2D3FEab296a9fba4e5FaDD2f7eD0 1okt --fees 1okt -b block -y
