package models

type PdePoolPairs struct {
	BeaconHeight    uint64 `json:"beaconheight" bson:"beaconheight"`
	Token1IDStr     string `json:"token1idstr" bson:"token1idstr"`
	Token1PoolValue uint64 `json:"token1poolvalue" bson:"token1poolvalue"`
	Token2IDStr     string `json:"token2idstr" bson:"token2idstr"`
	Token2PoolValue uint64 `json:"token2poolvalue" bson:"toke2poolvalue"`
}

type PdeTradeHistory struct {
	FromTokenIDStr   string `json:"fromtokenidstr" bson:"fromtokenidstr"`
	ToTokenIDStr     string `json:"totokenidstr" bson:"totokenidstr"`
	TraderAddressStr string `json:"traderaddressStr" bson:"traderaddressStr"`
	TradeAmount      uint64 `json:"tradeamount" bson:"tradeamount"`
	TradeFee         uint64 `json:"tradefee" bson:"tradefee"`
	ReceiveAmount    uint64 `json:"receiveamount" bson:"receiveamount"`
	ShardID          int    `json:"shardid" bson:"shardid"`
	RequestedTxID    string `json:"requestedtxid" bson:"requestedtxid"`
	ResponseTxID	 string `json:"responsetxid" bson:"responsetxid"`
	Type             int    `json:"type" bson:"type"`
	Status           string `json:"status" bson:"status"`
	BlockHeight      uint64 `json:"blockheight" bson:"blockheight"`
	LockTime         string `json:"locktime" bson:"locktime"`
}
