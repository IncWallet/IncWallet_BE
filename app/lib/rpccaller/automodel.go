package rpccaller

import (
	"incwallet/app/models"
	"time"
)

type AutoListToken struct {
	ID     int `json:"Id"`
	Result struct {
		ListCustomToken []struct {
			ID                 string        `json:"ID"`
			Name               string        `json:"Name"`
			Symbol             string        `json:"Symbol"`
			Image              string        `json:"Image"`
			Amount             uint64        `json:"Amount"`
			IsPrivacy          bool          `json:"IsPrivacy"`
			IsBridgeToken      bool          `json:"IsBridgeToken"`
			ListTxs            []interface{} `json:"ListTxs"`
			CountTxs           uint64        `json:"CountTxs"`
			InitiatorPublicKey string        `json:"InitiatorPublicKey"`
			TxInfo             string        `json:"TxInfo"`
		} `json:"ListCustomToken"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoTxHashHistory struct {
	ID      int                 `json:"Id"`
	Result  map[string][]string `json:"Result"`
	Error   interface{}         `json:"Error"`
	Params  []string            `json:"Params"`
	Method  string              `json:"Method"`
	Jsonrpc string              `json:"Jsonrpc"`
}

type AutoRandomCommitments struct {
	ID     int `json:"Id"`
	Result struct {
		CommitmentIndices  []uint64 `json:"CommitmentIndices"`
		MyCommitmentIndexs []uint64 `json:"MyCommitmentIndexs"`
		Commitments        []string `json:"Commitments"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoHasSerialNumber struct {
	ID      int           `json:"Id"`
	Result  []bool        `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

//type AutoCoinInfo struct {
//	PublicKey            string `json:"PublicKey"`
//	CoinCommitment       string `json:"CoinCommitment"`
//	SNDerivator          string `json:"SNDerivator"`
//	SerialNumber         string `json:"SerialNumber"`
//	Randomness           string `json:"Randomness"`
//	Value                string `json:"Value"`
//	Info                 string `json:"Info"`
//	CoinDetailsEncrypted string `json:"CoinDetailsEncrypted"`
//	TokenID              string `json:"TokenID"`
//}

type AutoListOutputPRV struct {
	ID     int `json:"Id"`
	Result struct {
		Outputs map[string][]models.Coins `json:"Outputs"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoListCustomToken struct {
	ID     int `json:"Id"`
	Result struct {
		ListCustomToken []struct {
			ID                 string        `json:"ID"`
			Name               string        `json:"Name"`
			Symbol             string        `json:"Symbol"`
			Image              string        `json:"Image"`
			Amount             int64         `json:"Amount"`
			IsPrivacy          bool          `json:"IsPrivacy"`
			IsBridgeToken      bool          `json:"IsBridgeToken"`
			ListTxs            []interface{} `json:"ListTxs"`
			CountTxs           int           `json:"CountTxs"`
			InitiatorPublicKey string        `json:"InitiatorPublicKey"`
			TxInfo             string        `json:"TxInfo"`
		} `json:"ListCustomToken"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoSendRawTxResult struct {
	ID     int `json:"Id"`
	Result struct {
		Base58CheckData string `json:"Base58CheckData"`
		TxID            string `json:"TxID"`
		ShardID         int    `json:"ShardID"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  []string    `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}

type AutoSendRawTokenTxResult struct {
	ID     int `json:"Id"`
	Result struct {
		Base58CheckData string `json:"Base58CheckData"`
		ShardID         int    `json:"ShardID"`
		TxID            string `json:"TxID"`
		TokenID         string `json:"TokenID"`
		TokenName       string `json:"TokenName"`
		TokenAmount     int    `json:"TokenAmount"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  []string    `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}

type AutoListAppTokenInfo struct {
	Result []struct {
		ID                 int         `json:"ID"`
		CreatedAt          time.Time   `json:"CreatedAt"`
		UpdatedAt          time.Time   `json:"UpdatedAt"`
		DeletedAt          interface{} `json:"DeletedAt"`
		TokenID            string      `json:"TokenID"`
		Symbol             string      `json:"Symbol"`
		OriginalSymbol     string      `json:"OriginalSymbol"`
		Name               string      `json:"Name"`
		ContractID         string      `json:"ContractID"`
		Decimals           int         `json:"Decimals"`
		PDecimals          int         `json:"PDecimals"`
		Status             int         `json:"Status"`
		Type               int         `json:"Type"`
		CurrencyType       int         `json:"CurrencyType"`
		PSymbol            string      `json:"PSymbol"`
		Default            bool        `json:"Default"`
		UserID             int         `json:"UserID"`
		PriceUsd           float64     `json:"PriceUsd"`
		Verified           bool        `json:"Verified"`
		LiquidityReward    float64     `json:"LiquidityReward"`
		PercentChange1H    string      `json:"PercentChange1h"`
		PercentChangePrv1H string      `json:"PercentChangePrv1h"`
		CurrentPrvPool     int64       `json:"CurrentPrvPool"`
		PricePrv           float64     `json:"PricePrv"`
		Volume24           int         `json:"volume24"`
	} `json:"Result"`
	Error interface{} `json:"Error"`
}

type AutoBeaconBestState struct {
	ID     int `json:"Id"`
	Result struct {
		BestBlockHash         string `json:"BestBlockHash"`
		PreviousBestBlockHash string `json:"PreviousBestBlockHash"`
		BestShardHash         map[string]string `json:"BestShardHash"`
		BestShardHeight 	  map[string]uint64 `json:"BestShardHeight"`
		Epoch                                  uint64           `json:"Epoch"`
		BeaconHeight                           uint64           `json:"BeaconHeight"`
		BeaconProposerIndex                    int           `json:"BeaconProposerIndex"`
		BeaconCommittee                        []string      `json:"BeaconCommittee"`
		BeaconPendingValidator                 []string `json:"BeaconPendingValidator"`
		CandidateShardWaitingForCurrentRandom  []string `json:"CandidateShardWaitingForCurrentRandom"`
		CandidateBeaconWaitingForCurrentRandom []string `json:"CandidateBeaconWaitingForCurrentRandom"`
		CandidateShardWaitingForNextRandom     []string `json:"CandidateShardWaitingForNextRandom"`
		CandidateBeaconWaitingForNextRandom    []string `json:"CandidateBeaconWaitingForNextRandom"`
		RewardReceiver                         struct {
		} `json:"RewardReceiver"`
		ShardCommittee 			map[string][]string `json:"ShardCommittee"`
		ShardPendingValidator 	map[string][]string `json:"ShardPendingValidator"`
		AutoStaking				map[string] bool`json:"AutoStaking"`
		CurrentRandomNumber    int64 `json:"CurrentRandomNumber"`
		CurrentRandomTimeStamp int   `json:"CurrentRandomTimeStamp"`
		IsGetRandomNumber      bool  `json:"IsGetRandomNumber"`
		MaxBeaconCommitteeSize int   `json:"MaxBeaconCommitteeSize"`
		MinBeaconCommitteeSize int   `json:"MinBeaconCommitteeSize"`
		MaxShardCommitteeSize  int   `json:"MaxShardCommitteeSize"`
		MinShardCommitteeSize  int   `json:"MinShardCommitteeSize"`
		ActiveShards           int   `json:"ActiveShards"`
		LastCrossShardState    struct {
		} `json:"LastCrossShardState"`
		ShardHandle interface{} `json:"ShardHandle"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoRewardAmount struct {
	ID     int `json:"Id"`
	Result map[string] map[string]uint64 `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoBeaconHeight struct {
	ID     int `json:"Id"`
	Result struct{
		ChainName string `json:"ChainName"`
		BestBlock map[string] struct{
			Height              uint64    `json:"Height"`
			Hash                string `json:"Hash"`
			TotalTxs            int    `json:"TotalTxs"`
			BlockProducer       string `json:"BlockProducer"`
			ValidationData      string `json:"ValidationData"`
			Epoch               uint64    `json:"Epoch"`
			Time                int    `json:"Time"`
			RemainingBlockEpoch uint64    `json:"RemainingBlockEpoch"`
			EpochBlock          int    `json:"EpochBlock"`
		} `json:"BestBlocks"`
		ActiveShards int `json:"ActiveShards"`
	} `json:"Result"`
	Error   interface{}   `json:"Error"`
	Params  []interface{} `json:"Params"`
	Method  string        `json:"Method"`
	Jsonrpc string        `json:"Jsonrpc"`
}

type AutoPdePoolPairs struct {
	Token1IDStr     string `json:"Token1IDStr"`
	Token1PoolValue uint64    `json:"Token1PoolValue"`
	Token2IDStr     string `json:"Token2IDStr"`
	Token2PoolValue uint64    `json:"Token2PoolValue"`
}

type AutoPdeState struct {
	ID     int `json:"Id"`
	Result struct {
		WaitingPDEContributions map[string] struct{
			ContributorAddressStr string `json:"ContributorAddressStr"`
			TokenIDStr            string `json:"TokenIDStr"`
			Amount                int64  `json:"Amount"`
			TxReqID               string `json:"TxReqID"`
		} `json:"WaitingPDEContributions"`
		PDEPoolPairs map[string]models.PdePoolPairs`json:"PDEPoolPairs"`
		PDEShares map[string] uint64 `json:"PDEShares"`
		PDETradingFees struct {} `json:"PDETradingFees"`
		BeaconTimeStamp int `json:"BeaconTimeStamp"`
	} `json:"Result"`
	Error  interface{} `json:"Error"`
	Params []struct {
		BeaconHeight int `json:"BeaconHeight"`
	} `json:"Params"`
	Method  string `json:"Method"`
	Jsonrpc string `json:"Jsonrpc"`
}

type AutoPdeTradeHistory struct {
	ID     int `json:"Id"`
	Result struct {
		PDEContributions []interface{} `json:"PDEContributions"`
		PDETrades        []models.PdeTradeHistory `json:"PDETrades"`
		PDEWithdrawals  []interface{} `json:"PDEWithdrawals"`
		BeaconTimeStamp int           `json:"BeaconTimeStamp"`
	} `json:"Result"`
	Error  interface{} `json:"Error"`
	Params []struct {
		BeaconHeight int `json:"BeaconHeight"`
	} `json:"Params"`
	Method  string `json:"Method"`
	Jsonrpc string `json:"Jsonrpc"`
}

type AutoBestBlockDetail struct {
	Height              uint64    `json:"Height"`
	Hash                string `json:"Hash"`
	TotalTxs            int    `json:"TotalTxs"`
	BlockProducer       string `json:"BlockProducer"`
	ValidationData      string `json:"ValidationData"`
	Epoch               int    `json:"Epoch"`
	Time                int    `json:"Time"`
	RemainingBlockEpoch int    `json:"RemainingBlockEpoch"`
	EpochBlock          int    `json:"EpochBlock"`
}

type AutoBestBlock struct {
	ID     int `json:"Id"`
	Result struct {
		BestBlocks map[string]AutoBestBlockDetail `json:"BestBlocks"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  string      `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}

type AutoRetrieveBlock struct {
	ID     int `json:"Id"`
	Result struct {
		Hash              string        `json:"Hash"`
		ShardID           int           `json:"ShardID"`
		Height            int           `json:"Height"`
		Confirmations     int           `json:"Confirmations"`
		Version           int           `json:"Version"`
		TxRoot            string        `json:"TxRoot"`
		Time              int           `json:"Time"`
		PreviousBlockHash string        `json:"PreviousBlockHash"`
		NextBlockHash     string        `json:"NextBlockHash"`
		TxHashes          []string `json:"TxHashes"`
		Txs               interface{}   `json:"Txs"`
		BlockProducer     string        `json:"BlockProducer"`
		ValidationData    string        `json:"ValidationData"`
		ConsensusType     string        `json:"ConsensusType"`
		Data              string        `json:"Data"`
		BeaconHeight      int           `json:"BeaconHeight"`
		BeaconBlockHash   string        `json:"BeaconBlockHash"`
		Round             int           `json:"Round"`
		Epoch             int           `json:"Epoch"`
		Reward            int           `json:"Reward"`
		RewardBeacon      int           `json:"RewardBeacon"`
		Fee               int           `json:"Fee"`
		Size              int           `json:"Size"`
		Instruction       []interface{} `json:"Instruction"`
		CrossShardBitMap  []interface{} `json:"CrossShardBitMap"`
	} `json:"Result"`
	Error   interface{} `json:"Error"`
	Params  []string    `json:"Params"`
	Method  string      `json:"Method"`
	Jsonrpc string      `json:"Jsonrpc"`
}
