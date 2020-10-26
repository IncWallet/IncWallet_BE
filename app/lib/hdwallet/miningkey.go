package hdwallet

import (
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"golang.org/x/crypto/sha3"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"math/big"
)

type CommitteePublicKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

func HashSHA3(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}

func HashKeccak256(data []byte) []byte {
	hashMachine := sha3.NewLegacyKeccak256()
	hashMachine.Write(data)
	return hashMachine.Sum(nil)
}

func GenerateKeySet(seed string, paymentAddress string) (*CommitteePublicKey, error) {
	seedPulicKey, err := Base58CheckDeserialize(paymentAddress)
	seedBytes, _, err := base58.Base58Check{}.Decode(seed)
	if err != nil {
		return nil, err
	}
	privKey := big.NewInt(0)
	privKey.SetBytes(HashSHA3(seedBytes))
	for {
		if privKey.Cmp(bn256.Order) == -1 {
			break
		}
		privKey.SetBytes(HashKeccak256(privKey.Bytes()))
	}

	pubKey := new(bn256.G2)
	pubKey = pubKey.ScalarBaseMult(privKey)

	committeePublicKey := new(CommitteePublicKey)
	committeePublicKey.IncPubKey = seedPulicKey.KeySet.PaymentAddress.Pk
	committeePublicKey.MiningPubKey = map[string][]byte{}
	_, blsPubKey := blsmultisig.KeyGen(seedBytes)
	blsPubKeyBytes := blsmultisig.PKBytes(blsPubKey)
	committeePublicKey.MiningPubKey[common.BlsConsensus] = blsPubKeyBytes
	_, briPubKey := bridgesig.KeyGen(seedBytes)
	briPubKeyBytes := bridgesig.PKBytes(&briPubKey)
	committeePublicKey.MiningPubKey[common.BridgeConsensus] = briPubKeyBytes
	return committeePublicKey, nil
}
