package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// DefaultGovAuthority is set to the gov module address.
// Extension point for chains to overwrite the default
var DefaultGovAuthority = sdk.AccAddress(address.Module("gov"))

func SubmitProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "submit-proposal",
		Short:        "Submit a mesh security proposal.",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		ProposalSetVirtualStakingMaxCapCmd(),
	)
	return cmd
}

func ProposalSetVirtualStakingMaxCapCmd() *cobra.Command {
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	cmd := &cobra.Command{
		Use:   "set-virtual-staking-max-cap [contract_addr_bech32] [max_cap] --title [text] --summary [text] --authority [address] --expedited [bool]",
		Short: "Submit a set virtual staking max cap proposal",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a proposal to set virtual staking maximum cap limit to the given contract.

Example:
$ %s tx meshsecurity submit-proposal set-virtual-staking-max-cap %s1l94ptufswr6v7qntax4m7nvn3jgf6k4gn2rknq 100stake --title "a title" --summary "a summary" --authority %s
`, version.AppName, bech32Prefix, DefaultGovAuthority.String())),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, metadata, deposit, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}

			expedited, err := cmd.Flags().GetBool(flagExpedited)
			if err != nil {
				return fmt.Errorf("expedited: %s", err)
			}

			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			src, err := parseSetVirtualStakingMaxCapArgs(args, authority)
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&src}, deposit, clientCtx.GetFromAddress().String(), metadata, proposalTitle, summary, expedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func parseSetVirtualStakingMaxCapArgs(args []string, authority string) (types.MsgSetVirtualStakingMaxCap, error) {
	maxCap, err := sdk.ParseCoinNormalized(args[1])
	if err != nil {
		return types.MsgSetVirtualStakingMaxCap{}, errorsmod.Wrap(err, "max cap")
	}

	msg := types.MsgSetVirtualStakingMaxCap{
		Authority: authority,
		Contract:  args[0],
		MaxCap:    maxCap,
	}
	return msg, nil
}

func addCommonProposalFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(cli.FlagTitle, "", "Title of proposal")
	cmd.Flags().String(cli.FlagSummary, "", "Summary of proposal")
	cmd.Flags().String(cli.FlagMetadata, "", "Metadata of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "Deposit of proposal")
	cmd.Flags().String(flagAuthority, DefaultGovAuthority.String(), "The address of the governance account. Default is the sdk gov module account")
}

func getProposalInfo(cmd *cobra.Command) (client.Context, string, string, string, sdk.Coins, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return client.Context{}, "", "", "", nil, err
	}

	proposalTitle, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return clientCtx, proposalTitle, "", "", nil, err
	}

	summary, err := cmd.Flags().GetString(cli.FlagSummary)
	if err != nil {
		return client.Context{}, proposalTitle, summary, "", nil, err
	}

	metadata, err := cmd.Flags().GetString(cli.FlagMetadata)
	if err != nil {
		return client.Context{}, proposalTitle, summary, metadata, nil, err
	}

	depositArg, err := cmd.Flags().GetString(cli.FlagDeposit)
	if err != nil {
		return client.Context{}, proposalTitle, summary, metadata, nil, err
	}

	deposit, err := sdk.ParseCoinsNormalized(depositArg)
	if err != nil {
		return client.Context{}, proposalTitle, summary, metadata, deposit, err
	}

	return clientCtx, proposalTitle, summary, metadata, deposit, nil
}
