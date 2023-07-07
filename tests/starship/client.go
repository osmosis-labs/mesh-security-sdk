package e2e

import (
	"fmt"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/go-bip39"
	"os"

	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

type Client struct {
	Logger *zap.Logger
	Config *starship.Config

	Name        string
	Address     string
	ChainID     string
	ChainConfig *lens.ChainClientConfig
	Client      *lens.ChainClient
}

func NewClient(name string, logger *zap.Logger, starshipClient *starship.ChainClient, modulesBasics []module.AppModuleBasic) (*Client, error) {
	chainClient := &Client{
		Name:    name,
		Logger:  logger,
		Config:  starshipClient.Config,
		ChainID: starshipClient.ChainID,
	}

	// fetch chain registry from the local registry
	registry, err := starshipClient.GetChainRegistry()
	if err != nil {
		return nil, err
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
