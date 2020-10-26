package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"strconv"
)

type Metadata interface {
	GetType() int
	SetType(int)
	Hash() *common.Hash
}

//trade request metadata
type PDETradeRequest struct {
	TokenIDToBuyStr     string
	TokenIDToSellStr    string
	SellAmount          uint64 // must be equal to vout value
	MinAcceptableAmount uint64
	TradingFee          uint64
	TraderAddressStr    string
	Type				int
}
type PDETradeResponse struct {
	Type	      int
	TradeStatus   string
	RequestedTxID string
}

func (pdeTradeRequest PDETradeRequest) GetType() int {return pdeTradeRequest.Type}
func (pdeTradeRequest *PDETradeRequest) SetType(iType int) {pdeTradeRequest.Type = iType}
func (pdeTradeRequest PDETradeRequest) Hash() *common.Hash {
	tmp := strconv.Itoa(pdeTradeRequest.Type)
	data := []byte(tmp)
	hash := common.HashH(data)
	record := hash.String()
	record += pdeTradeRequest.TokenIDToBuyStr
	record += pdeTradeRequest.TokenIDToSellStr
	record += pdeTradeRequest.TraderAddressStr
	record += strconv.FormatUint(pdeTradeRequest.SellAmount, 10)
	record += strconv.FormatUint(pdeTradeRequest.MinAcceptableAmount, 10)
	record += strconv.FormatUint(pdeTradeRequest.TradingFee, 10)
	finalHash := common.HashH([]byte(record))
	return &finalHash
}

func NewPDETradeRequestMetadata(
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee          uint64,
	traderAddressStr    string) *PDETradeRequest {
	return &PDETradeRequest{
		TokenIDToBuyStr: tokenIDToBuyStr,
		TokenIDToSellStr: tokenIDToSellStr,
		SellAmount: sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee: tradingFee,
		TraderAddressStr: traderAddressStr,
		Type: PDETradeRequestMeta,
	}
}

func NewPDETradeCrossRequestMetadata(
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	tradingFee          uint64,
	traderAddressStr    string) *PDETradeRequest {
	return &PDETradeRequest{
		TokenIDToBuyStr: tokenIDToBuyStr,
		TokenIDToSellStr: tokenIDToSellStr,
		SellAmount: sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		TradingFee: tradingFee,
		TraderAddressStr: traderAddressStr,
		Type: PDECrossPoolTradeRequestMeta,
	}
}

type StopAutoStakingMetadata struct {
	Type	      int
	CommitteePublicKey string
}


func (stopStaking StopAutoStakingMetadata) GetType() int {return stopStaking.Type}
func (stopStaking *StopAutoStakingMetadata) SetType(iType int) {stopStaking.Type = iType}
func (stopStaking StopAutoStakingMetadata) Hash() *common.Hash {
	tmp := strconv.Itoa(stopStaking.Type)
	data := []byte(tmp)
	hash := common.HashH(data)
	return &hash
}
func NewStopAutoStakingMetadata(committeePublicKey string) (*StopAutoStakingMetadata) {
	return &StopAutoStakingMetadata{
		Type:       StopAutoStakingMeta,
		CommitteePublicKey: committeePublicKey,
	}
}
