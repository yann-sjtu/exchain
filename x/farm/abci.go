package farm

import (
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	abci "github.com/okex/exchain/libs/tendermint/abci/types"
	"github.com/okex/exchain/x/farm/keeper"
	"github.com/okex/exchain/x/farm/types"
)

// BeginBlocker allocates the native token to the pools in PoolsYieldNativeToken
// according to the value of locked token in pool
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

}

// calculateAllocateInfo gets all pools in PoolsYieldNativeToken
func calculateAllocateInfo(ctx sdk.Context, k keeper.Keeper) (map[string]sdk.Dec, []types.FarmPool, sdk.Dec) {
	lockedPoolValue := make(map[string]sdk.Dec)
	var pools types.FarmPools
	totalPoolsValue := sdk.ZeroDec()

	store := ctx.KVStore(k.StoreKey())
	iterator := sdk.KVStorePrefixIterator(store, types.PoolsYieldNativeTokenPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		poolName := types.SplitPoolsYieldNativeTokenKey(iterator.Key())
		pool, found := k.GetFarmPool(ctx, poolName)
		if !found {
			panic("should not happen")
		}
		poolValue := k.GetPoolLockedValue(ctx, pool)
		if poolValue.LTE(sdk.ZeroDec()) {
			continue
		}
		pools = append(pools, pool)
		lockedPoolValue[poolName] = poolValue
		totalPoolsValue = totalPoolsValue.Add(poolValue)
	}
	return lockedPoolValue, pools, totalPoolsValue
}
