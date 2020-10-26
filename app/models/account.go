package models

/*
Account models
*/
type Account struct {
	Name           string     `json:"name" bson:"name"`
	Index          uint32     `json:"index" bson:"index"`
	PublicKey      string     `json:"publickey" bson:"publickey"`
	PaymentAddress string     `json:"paymentaddress" bson:"paymentaddress"`
	ViewKey        string     `json:"viewkey" bson:"viewkey"`
	MiningKey      string     `json:"miningkey" bson:"miningkey"`
	Wallet         string     `json:"wallet" bson:"wallet"`
	Crypto         CryptoJSON `json:"crypto" bson:"crypto"`
}

/*
Address book models
*/
type AddressBook struct {
	Name           string `json:"name" bson:"name"`
	PaymentAddress string `json:"paymentaddress" bson:"paymentaddress"`
	ChainName      string `json:"chainname" bson:"chainname"`
	ChainType      string `json:"chaintype" bson:"chaintype"`
}
