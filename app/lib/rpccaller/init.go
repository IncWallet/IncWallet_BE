package rpccaller

import (
	"incwallet/app/lib/common"
)

type RPCService struct {
	Network string
	Url     string
}

func (rpcCaller *RPCService) Init(name string) {
	if name == common.Mainnet {
		rpcCaller.Url = common.URLMainnet
		rpcCaller.Network = name
	}
	if name == common.Testnet {
		rpcCaller.Url = common.URLTestnet
		rpcCaller.Network = name
	}
	if name == common.Local {
		rpcCaller.Url = common.URLLocal
		rpcCaller.Network = name
	}

}

func (rpcCaller *RPCService) InitMainnet(url string) {
	rpcCaller.Url = url
	rpcCaller.Network = common.Mainnet
}

func (rpcCaller *RPCService) InitTestnet(url string) {
	rpcCaller.Url = url
	rpcCaller.Network = common.Testnet
}

func (rpcCaller *RPCService) InitLocal(url string) {
	rpcCaller.Url = url
	rpcCaller.Network = common.Local
}
