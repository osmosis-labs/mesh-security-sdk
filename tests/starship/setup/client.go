package setup

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	pb "github.com/cosmology-tech/starship/registry/registry"
	"github.com/cosmos/go-bip39"
	lens "github.com/strangelove-ventures/lens/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Client struct {
	Logger *zap.Logger
	Config *starship.Config

	Name           string
	Address        string
	ChainID        string
	Denom          string
	StarshipClient *starship.ChainClient
	ChainConfig    *lens.ChainClientConfig
	Client         *lens.ChainClient
}

func NewClient(name string, logger *zap.Logger, starshipClient *starship.ChainClient, modulesBasics []module.AppModuleBasic) (*Client, error) {
	// fetch Chain registry from the local registry
	registry, err := starshipClient.GetChainRegistry()
	if err != nil {
		return nil, err
	}

	chainClient := &Client{
		Name:           name,
		Logger:         logger,
		Config:         starshipClient.Config,
		ChainID:        starshipClient.ChainID,
		Denom:          registry.Fees.FeeTokens[0].Denom,
		StarshipClient: starshipClient,
	}

	ccc := &lens.ChainClientConfig{
		ChainID:        starshipClient.ChainID,
		RPCAddr:        starshipClient.GetRPCAddr(),
		KeyringBackend: "test",
		Debug:          true,
		Timeout:        "20s",
		SignModeStr:    "direct",
		AccountPrefix:  *registry.Bech32Prefix,
		GasAdjustment:  1.5,
		GasPrices:      fmt.Sprintf("%f%s", registry.Fees.FeeTokens[0].HighGasPrice, registry.Fees.FeeTokens[0].Denom),
		MinGasAmount:   10000,
		Slip44:         int(registry.Slip44),
		Modules:        modulesBasics,
	}

	client, err := lens.NewChainClient(logger, ccc, os.Getenv("HOME"), os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	chainClient.ChainConfig = ccc
	chainClient.Client = client

	err = chainClient.Initialize()
	if err != nil {
		return nil, err
	}

	// credit amount from faucet
	err = starshipClient.CreditFromFaucet(chainClient.Address)
	if err != nil {
		return nil, err
	}

	return chainClient, nil
}

func (c *Client) GetChainID() string {
	return c.ChainID
}

func (c *Client) GetChainDenom() (string, error) {
	return c.Denom, nil
}

func (c *Client) GetHeight() (int64, error) {
	status, err := c.Client.RPCClient.Status(context.Background())
	if err != nil {
		return -1, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

func (c *Client) Initialize() error {
	keyName := fmt.Sprintf("client-%s-%s", c.ChainID, c.Name)

	wallet, err := c.CreateRandWallet(keyName)
	if err != nil {
		return err
	}

	c.Address = wallet
	c.ChainConfig.Key = keyName

	return nil
}

func (c *Client) CreateRandWallet(keyName string) (string, error) {
	// delete key if already exists
	_ = c.Client.DeleteKey(keyName)

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	address, err := c.CreateWallet(keyName, mnemonic)
	if err != nil {
		return "", err
	}

	return address, nil
}

func (c *Client) CreateWallet(keyName, mnemonic string) (string, error) {
	// delete key if already exists
	_ = c.Client.DeleteKey(keyName)

	walletAddr, err := c.Client.RestoreKey(keyName, mnemonic, uint32(c.ChainConfig.Slip44))
	if err != nil {
		return "", err
	}

	return walletAddr, nil
}

// StakeTokens self stakes tokens from the client address
func (c *Client) StakeTokens(valAddr string, amount int, denom string) error {
	stakeMsg := &stakingtypes.MsgDelegate{
		DelegatorAddress: c.Address,
		ValidatorAddress: valAddr,
		Amount:           sdk.NewInt64Coin(denom, int64(amount)),
	}

	_, err := c.Client.SendMsg(context.Background(), stakeMsg, "")
	if err != nil {
		return err
	}

	return err
}

// WaitForTx will wait for the tx to complete, fail if not able to find tx
func (c *Client) WaitForTx(t *testing.T, txHex string) {
	var tx *coretypes.ResultTx
	var err error
	require.Eventuallyf(t,
		func() bool {
			tx, err = c.Client.QueryTx(context.Background(), txHex, false)
			if err != nil {
				return false
			}
			if tx.TxResult.Code == 0 {
				return true
			}
			return false
		},
		300*time.Second,
		time.Second,
		"waited for too long, still txn not successful",
	)
	require.NotNil(t, tx)
}

// WaitForHeight will wait till the Chain reaches the block height
func (c *Client) WaitForHeight(t *testing.T, height int64) {
	require.Eventuallyf(t,
		func() bool {
			curHeight, err := c.GetHeight()
			assert.NoError(t, err)
			if curHeight >= height {
				return true
			}
			return false
		},
		300*time.Second,
		5*time.Second,
		"waited for too long, still height did not reach desired block height",
	)
}

// GetIBCInfo will fetch Fetch IBC info from chain registry
func (c *Client) GetIBCInfo(chain2 string) (*pb.IBCData, error) {
	return c.StarshipClient.GetIBCInfo(chain2)
}

func (c *Client) GetIBCChannel(chain2 string) (*pb.ChannelData, error) {
	return c.StarshipClient.GetIBCChannel(chain2)
}
