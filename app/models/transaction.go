package models

type TxHistory struct {
	TxHash     string `json:"txhash" bson:"txhash"`
	PublicKey  string `json:"publickey" bson:"publickey"`
	LockTime   string `json:"locktime" bson:"locktime"`
	Type       string `json:"type" bson:"type"`
	Fee        uint64 `json:"fee" bson:"fee"`
	VInPRVs    uint64 `json:"vinprvs" bson:"vinprvs"`
	VOutPRVs   uint64 `json:"voutprvs" bson:"voutprvs"`
	TokenID    string `json:"tokenid" bson:"tokenid"`
	TokenFee   uint64 `json:"tokenfee" bson:"tokenfee"`
	VInTokens  uint64 `json:"vintokens" bson:"vintokens"`
	VOutTokens uint64 `json:"vouttokens" bson:"vouttokens"`
}

type AutoTxByHash struct {
	ID     int `json:"Id"`
	Result struct {
		BlockHash   string `json:"BlockHash"`
		BlockHeight uint64 `json:"BlockHeight"`
		TxSize      int    `json:"TxSize"`
		Index       int    `json:"Index"`
		ShardID     int    `json:"ShardID"`
		Hash        string `json:"Hash"`
		Version     int    `json:"Version"`
		Type        string `json:"Type"`
		LockTime    string `json:"LockTime"`
		Fee         uint64 `json:"Fee"`
		Image       string `json:"Image"`
		IsPrivacy   bool   `json:"IsPrivacy"`
		Proof       string `json:"Proof"`
		ProofDetail struct {
			InputCoins  []AutoCoin `json:"InputCoins"`
			OutputCoins []AutoCoin `json:"OutputCoins"`
		} `json:"ProofDetail"`
		InputCoinPubKey               string `json:"InputCoinPubKey"`
		SigPubKey                     string `json:"SigPubKey"`
		Sig                           string `json:"Sig"`
		Metadata                      string `json:"Metadata"`
		CustomTokenData               string `json:"CustomTokenData"`
		PrivacyCustomTokenID          string `json:"PrivacyCustomTokenID"`
		PrivacyCustomTokenName        string `json:"PrivacyCustomTokenName"`
		PrivacyCustomTokenSymbol      string `json:"PrivacyCustomTokenSymbol"`
		PrivacyCustomTokenData        string `json:"PrivacyCustomTokenData"`
		PrivacyCustomTokenProofDetail struct {
			InputCoins  []AutoCoin `json:"InputCoins"`
			OutputCoins []AutoCoin `json:"OutputCoins"`
		} `json:"PrivacyCustomTokenProofDetail"`
		PrivacyCustomTokenIsPrivacy bool   `json:"PrivacyCustomTokenIsPrivacy"`
		PrivacyCustomTokenFee       uint64 `json:"PrivacyCustomTokenFee"`
		IsInMempool                 bool   `json:"IsInMempool"`
		IsInBlock                   bool   `json:"IsInBlock"`
		Info                        string `json:"Info"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  []string    `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}

type AutoCoin struct {
	CoinDetails struct {
		PublicKey      string   `json:"PublicKey"`
		CoinCommitment string   `json:"CoinCommitment"`
		SNDerivator    struct{} `json:"SNDerivator"`
		SerialNumber   string   `json:"SerialNumber"`
		Randomness     struct{} `json:"Randomness"`
		Value          uint64   `json:"Value"`
		Info           string   `json:"Info"`
	} `json:"CoinDetails"`
	CoinDetailsEncrypted string `json:"CoinDetailsEncrypted"`
}
