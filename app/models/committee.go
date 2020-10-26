package models

/*
Committee models
*/
type Committee struct {
	Epoch   uint64 `json:"epoch"`
	BeaconHeight uint64 `json:"beaconheight" bson:"beaconheight"`
	Key     string `json:"key" bson:"key"`
	Role    string `json:"role" bson:"role"`
	ShardId int    `json:"shardid" bson:"sharid"`
	Index   int    `json:"index" bson:"index"`
}
type CommitteReward struct {
	BeaconHeight uint64 `json:"beaconheight" bson:"beaconheight"`
	PublicKey string `json:"publickey" bson:"publickey"`
	Amount    uint64 `json:"amount" bson:"amount"`
}
