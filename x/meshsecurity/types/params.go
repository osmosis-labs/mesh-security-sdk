package types

// todo: should be generated from protobuf
// todo: also add new messages to update params
type Params struct {
	// todo: add safety net with one of:
	// - total max cap amount (for all contracts)
	// - max number of virtualstaking contract
}

// todo: implement proper
func (p Params) ValidateBasic() error {
	return nil
}
