package common

const (
	PriKeyType         = byte(0x0) // serialize wallet account key into string with only PRIVATE KEY of account keyset
	PaymentAddressType = byte(0x1) // serialize wallet account key into string with only PAYMENT ADDRESS of account keyset
	ReadonlyKeyType    = byte(0x2) // serialize wallet account key into string with only READONLY KEY of account keyset
)

const (
	CustomTokenInit = iota
	CustomTokenTransfer
)

const (
	SeedKeyLen     = 64 // bytes
	ChildNumberLen = 4  // bytes
	ChainCodeLen   = 32 // bytes

	PrivateKeySerializedLen = 108 // len string

	PrivKeySerializedBytesLen     = 75 // bytes
	PaymentAddrSerializedBytesLen = 71 // bytes
	ReadOnlyKeySerializedBytesLen = 71 // bytes

	PrivKeyBase58CheckSerializedBytesLen     = 107 // len string
	PaymentAddrBase58CheckSerializedBytesLen = 103 // len string
	ReadOnlyKeyBase58CheckSerializedBytesLen = 103 // len string
)

const (
	EmptyString       = ""
	ZeroByte          = byte(0x00)
	DateOutputFormat  = "2006-01-02T15:04:05.999999"
	BigIntSize        = 32 // bytes
	CheckSumLen       = 4  // bytes
	AESKeySize        = 32 // bytes
	Int32Size         = 4  // bytes
	Uint32Size        = 4  // bytes
	Uint64Size        = 8  // bytes
	HashSize          = 32 // bytes
	MaxHashStringSize = HashSize * 2
	Base58Version     = 0
)

const (
	MinUnspentCoins = 10
	MaxTxInput      = 31
)

const (
	PRVID     = "0000000000000000000000000000000000000000000000000000000000000004"
	USDTID    = "716fd1009e2a1669caacc36891e707bfdf02590f96ebd897548e8963c95ebac0"
	BTCID     = "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
	PRVSymbol = "PRV"
)

const (
	Mainnet = "mainnet"
	Testnet = "testnet"
	Local   = "local"
	ShardNumber = 8
	PDETradeResponse1 = 92
	PDETradeResponse2 = 206
)

const (
	//URLMainnet = "https://community-fullnode.incognito.org"
	//URLMainnet = "https://fullnode.incognito.org"
	//URLMainnet = "http://localhost:9334"
	URLMainnet = "http://167.86.99.232:9334"
	URLTestnet = "https://testnet.incognito.org/fullnode"
	URLLocal   = "http://localhost:9334"
)

const (
	BurnAddress1 = "15pABFiJVeh9D5uiQEhQX4SVibGGbdAVipQxBdxkmDqAJaoG1EdFKHBrNfs"
	BurnAddress2 = "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"
)

const (
	PointCompressed       byte = 0x2
	ElGamalCiphertextSize      = 64 // bytes
	SchnMultiSigSize           = 65 // bytes
)


const (
	NodeModeRelay  = "relay"
	NodeModeShard  = "shard"
	NodeModeAuto   = "auto"
	NodeModeBeacon = "beacon"

	BeaconRole     = "beacon"
	ShardRole      = "shard"
	CommitteeRole  = "committee"
	ProposerRole   = "proposer"
	ValidatorRole  = "validator"
	PendingRole    = "pending"
	CandidateRole  = "candidate"
	SyncingRole    = "syncing" //this is for shard case - when beacon tell it is committee, but its state not
	WaitingRole    = "waiting"
	MaxShardNumber = 8

	BlsConsensus    = "bls"
	BridgeConsensus = "dsa"
	IncKeyType      = "inc"
	ImageURLPrefix 	= "https://s3.amazonaws.com/incognito-org/wallet/cryptocurrency-icons/32@2x/color/"
	ImageURLSubfix  = "@2x.png"
)

const (
	GcmAdditionData = "incognito chain"
)

const (
	CryptoStoreVersion = 1
)

const MaxIndex = 4294967295

var BurnAddress1BytesDecode = []byte{1, 32, 99, 183, 246, 161, 68, 172, 228, 222, 153, 9, 172, 39, 208, 245, 167, 79, 11, 2, 114, 65, 241, 69, 85, 40, 193, 104, 199, 79, 70, 4, 53, 0, 0, 163, 228, 236, 208}

const (
	TransferTokenType = "tp"
	TransferPRVType   = "n"
	SalaryPRVType     = "s"

	SendStr              = "Send"
	ReceiveStr           = "Receive"
	TransferTokenTypeStr = "Token"
	TransferPRVTypeStr   = "PRV"
)

// size data for incognito key and signature
const (
	// for key size
	PrivateKeySize      = 32  // bytes
	PublicKeySize       = 32  // bytes
	BLSPublicKeySize    = 128 // bytes
	BriPublicKeySize    = 33  // bytes
	TransmissionKeySize = 32  //bytes
	ReceivingKeySize    = 32  // bytes
	PaymentAddressSize  = 64  // bytes
	// for signature size
	// it is used for both privacy and no privacy
	SigPubKeySize    = 32
	SigNoPrivacySize = 64
	SigPrivacySize   = 96
	IncPubKeyB58Size = 51
)

const (
	LocalURL = "http://localhost:9000"
	ServiceURL = "http://167.86.99.232"
)
