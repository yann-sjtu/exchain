package mint

import (
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	minter := keeper.GetMinterCustom(ctx)
	params := keeper.GetParams(ctx)
	return NewGenesisState(minter, params, keeper.GetOriginalMintedPerBlock())
}
