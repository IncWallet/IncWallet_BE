package models

import "gopkg.in/mgo.v2/bson"

/*
State models
*/
type State struct {
	ID           bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	WalletID     string        `json:"walletid" bson:"walletid"`
	AccountID    string        `json:"accountid" bson:"accountid"`
	ShardHeight  int           `json:"shardheight" bson:"shardheight"`
	BeaconHeight int           `json:"beaconheight" bson:"beaconheight"`
	Network      string        `json:"network" bson:"network"`
}

/*
Network model
*/
type Network struct {
	Name string `json:"name" bson:"name"`
	Url  string `json:"url" bson:"url"`
}
