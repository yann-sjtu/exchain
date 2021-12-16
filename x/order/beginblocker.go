package order

import (
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"

	"github.com/okex/exchain/x/order/keeper"
	//"github.com/okex/exchain/x/common/version"
)

// BeginBlocker runs the logic of BeginBlocker with version 0.
// BeginBlocker resets keeper cache.
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
}
