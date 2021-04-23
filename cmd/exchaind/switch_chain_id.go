package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/node"
	sm "github.com/tendermint/tendermint/state"
)

func switchChainIDCmd(ctx *server.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-chain",
		Short: "Switch chain ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(flags.FlagHome))

			_, stateDB, _, err := initDBs(config, node.DefaultDBProvider)
			if err != nil {
				return err
			}

			genesisDocProvider := node.DefaultGenesisDocProviderFunc(config)
			state, _, err := node.LoadStateFromDBOrGenesisDocProvider(stateDB, genesisDocProvider)
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", state)
			state.ChainID = viper.GetString(flags.FlagChainID)
			sm.SaveState(stateDB, state)

			return nil
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "Chain ID of tendermint nod")
	cmd.MarkFlagRequired(flags.FlagChainID)

	return cmd
}
