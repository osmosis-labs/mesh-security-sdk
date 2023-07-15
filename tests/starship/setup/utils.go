package setup

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
)

// IBCTransferTokens will transfer chain native token from chain1 to chain2 at given address
func IBCTransferTokens(chain1, chain2 *Client, chain2Addr string, amount int) error {
	channel, err := chain1.GetIBCChannel(chain2.GetChainID())
	if err != nil {
		return err
	}

	denom, err := chain1.GetChainDenom()
	if err != nil {
		return err
	}

	coin := sdk.Coin{Denom: denom, Amount: sdk.NewInt(int64(amount))}
	req := &transfertypes.MsgTransfer{
		SourcePort:       channel.Chain_2.PortId,
		SourceChannel:    channel.Chain_2.ChannelId,
		Token:            coin,
		Sender:           chain1.Address,
		Receiver:         chain2Addr,
		TimeoutHeight:    clienttypes.NewHeight(12300, 45600),
		TimeoutTimestamp: 0,
		Memo:             fmt.Sprintf("testsetup: transfer token from %s to %s", chain1.GetChainID(), chain2.GetChainID()),
	}

	res, err := chain1.Client.SendMsg(context.Background(), req, "")
	if err != nil {
		if res != nil {
			fmt.Printf("code: %d, logs: %v\n", res.Code, res.Logs)
		}
		fmt.Printf("error: %s\n", err)
		return err
	}

	return nil
}
