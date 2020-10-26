package models

type Wallet struct {
	WalletId  string     `json:"walletid" bson:"walletid"`
	ShardID	  int		 `json:"shardid" bson:"shardid"`
	Crypto    CryptoJSON `json:"crypto" bson:"crypto"`
	Version   int        `json:"version" bson:"version"`
	Network   string     `json:"network" bson:"network"`
	Timestamp int64      `json:"timestamp" bson:"timestamp"`
}

type CryptoJSON struct {
	CipherName   string       `json:"ciphername"`
	CipherText   string       `json:"ciphertext"`
	Nonce        string       `json:"nonce"`
	KDF          string       `json:"kdf"`
	ScryptParams ScryptParams `json:"scryptparams"`
}

type ScryptParams struct {
	N      int    `json:"n"`
	R      int    `json:"r"`
	P      int    `json:"p"`
	KeyLen int    `json:"keylen"`
	Salt   string `json:"salt"`
}
