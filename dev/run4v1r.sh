#!/usr/bin/env bash
DATAPATH="$HOME/Downloads"

db_backend="goleveldb"

while getopts "i" opt; do
  case $opt in
  i)
    NAME="exchaind"
    MYNAME="run4v1r.sh"
    ps -ef|grep "$NAME"|grep -v grep |grep -v $MYNAME |awk '{print "kill -9 "$2", "$8}'
    ps -ef|grep "$NAME"|grep -v grep |awk '{print "kill -9 "$2}' | sh
    echo "All <$NAME> killed!"

    echo "INIT"
    exchaind unsafe-reset-all --home ${DATAPATH}/cache/node0/exchaind
    exchaind unsafe-reset-all --home ${DATAPATH}/cache/node1/exchaind
    exchaind unsafe-reset-all --home ${DATAPATH}/cache/node2/exchaind
    exchaind unsafe-reset-all --home ${DATAPATH}/cache/node3/exchaind
    exchaind unsafe-reset-all --home ${DATAPATH}/cache/node4/exchaind
    ;;
  esac
done

sleep 2

mempool_size=10000
tpb=3000
timeout_commit="2000ms"

# val0 with seed
nohup exchaind start  --chain-id exchain-67 --log_level "state:info,main:info,root-multi:info" --db_backend $db_backend --pruning=everything \
                --prof_laddr="localhost:6061" \
                --mempool.max_tx_num_per_block=${tpb} --mempool.size=${mempool_size} \
                --rpc.unsafe --disable-abci-query-mutex=true --consensus.timeout_commit ${timeout_commit} \
                --p2p.seed_mode=true --p2p.allow_duplicate_ip  --p2p.pex=false --p2p.addr_book_strict=false --p2p.laddr tcp://127.0.0.1:20056 \
                --rest.laddr tcp://0.0.0.0:20059 --rpc.laddr tcp://0.0.0.0:20057 --home ${DATAPATH}/cache/node0/exchaind > ${DATAPATH}/cache/0.log &

# val1
nohup exchaind start  --chain-id exchain-67 --log_level "state:info,main:info,root-multi:info" --db_backend $db_backend --pruning=everything \
                --prof_laddr="localhost:6062" \
                --mempool.max_tx_num_per_block=${tpb} --mempool.size=${mempool_size} \
                --rpc.unsafe --disable-abci-query-mutex=true --consensus.timeout_commit ${timeout_commit} \
                --p2p.seed_mode=false --p2p.allow_duplicate_ip --p2p.pex=false --p2p.addr_book_strict=false --p2p.laddr tcp://127.0.0.1:20156 --p2p.seeds 0b066ca0790f27a6595560b23bf1a1193f100797@127.0.0.1:20056 \
                --rest.laddr tcp://0.0.0.0:20159 --rpc.laddr tcp://0.0.0.0:20157 --home ${DATAPATH}/cache/node1/exchaind > ${DATAPATH}/cache/1.log &

# val2
nohup exchaind start  --chain-id exchain-67 --log_level "state:info,main:info,root-multi:info" --db_backend $db_backend --pruning=nothing \
                --prof_laddr="localhost:6063" \
                --mempool.max_tx_num_per_block=${tpb} --mempool.size=${mempool_size} \
                --rpc.unsafe --disable-abci-query-mutex=true --consensus.timeout_commit ${timeout_commit} \
                --p2p.seed_mode=false --p2p.allow_duplicate_ip --p2p.pex=false --p2p.addr_book_strict=false --p2p.laddr tcp://127.0.0.1:20256 --p2p.seeds 0b066ca0790f27a6595560b23bf1a1193f100797@127.0.0.1:20056 \
                --rest.laddr tcp://0.0.0.0:20259 --rpc.laddr tcp://0.0.0.0:20257 --home ${DATAPATH}/cache/node2/exchaind > ${DATAPATH}/cache/2.log &

# val3
nohup exchaind start  --chain-id exchain-67 --log_level "state:info,main:info,root-multi:info" --db_backend $db_backend --pruning=everything \
                --prof_laddr="localhost:6064" \
                --mempool.max_tx_num_per_block=${tpb} --mempool.size=${mempool_size} \
                --rpc.unsafe --disable-abci-query-mutex=true --consensus.timeout_commit ${timeout_commit} \
                --p2p.seed_mode=false --p2p.allow_duplicate_ip --p2p.pex=false --p2p.addr_book_strict=false --p2p.laddr tcp://127.0.0.1:20356 --p2p.seeds 0b066ca0790f27a6595560b23bf1a1193f100797@127.0.0.1:20056 \
                --rest.laddr tcp://0.0.0.0:20359 --rpc.laddr tcp://0.0.0.0:20357 --home ${DATAPATH}/cache/node3/exchaind > ${DATAPATH}/cache/3.log &

# full0
nohup exchaind start  --chain-id exchain-67 --log_level "state:info,main:info,root-multi:info" --db_backend $db_backend --pruning=nothing \
                --mempool.max_tx_num_per_block=${tpb} --mempool.size=${mempool_size} \
                --local_perf=tx \
                --rpc.unsafe --disable-abci-query-mutex=true --consensus.timeout_commit ${timeout_commit} \
                --p2p.seed_mode=false --p2p.allow_duplicate_ip --p2p.pex=false --p2p.addr_book_strict=false --p2p.laddr tcp://127.0.0.1:20456 --p2p.seeds 0b066ca0790f27a6595560b23bf1a1193f100797@127.0.0.1:20056 \
                --rest.laddr tcp://localhost:8545 --rpc.laddr tcp://0.0.0.0:26657 --home ${DATAPATH}/cache/node4/exchaind > ${DATAPATH}/cache/4.log &

# --halt-height 150
# --local_perf=tx \