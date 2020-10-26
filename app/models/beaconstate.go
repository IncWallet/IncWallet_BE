package models
/*
BeaconState Models
*/
type BeaconState struct {
	Height uint64 `json:"height"`
	Epoch uint64 `json:"epoch"`
	RemainingBlockEpoch uint64 `json:"remainingblockepoch"`
}
