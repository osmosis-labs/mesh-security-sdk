package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Mesh security transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}

	txCmd.AddCommand(NewSetConsumerCommissionRateCmd())
	return txCmd
}

func NewSetConsumerCommissionRateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-consumer-commission-rate [consumer-chain-id] [commission-rate]",
		Short: "set a per-consumer chain commission",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Note that the "commission-rate" argument is a fraction and should be in the range [0,1].
			Example:
			%s set-consumer-commission-rate consumer-1 0.5 --from node0 --home ../node0`,
				version.AppName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			providerValAddr := clientCtx.GetFromAddress()

			commission, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgSetConsumerCommissionRate(args[0], commission, sdk.ValAddress(providerValAddr))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}
