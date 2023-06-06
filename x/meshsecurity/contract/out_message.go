package contract

type SudoMsg struct {
	Rebalance *struct{} `json:"rebalance"`
}
