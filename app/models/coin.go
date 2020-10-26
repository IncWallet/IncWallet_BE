package models

type Coins struct {
	PublicKey            string `json:"publickey" bson:"publickey"`
	CoinCommitment       string `json:"coincommitment" bson:"coincommitment"`
	SNDerivator          string `json:"snderivator" bson:"snderivator"`
	SerialNumber         string `json:"serialnumber" bson:"serialnumber"`
	Randomness           string `json:"randomness" bson:"randomness"`
	Value                string `json:"value" bson:"value"`
	Info                 string `json:"info" bson:"info"`
	CoinDetailsEncrypted string `json:"coindetailsencrypted" bson:"coindetailsencrypted"`
	IsSpent				 bool   `json:"isspent" bson:"isspent"`
	TokenID				 string `json:"tokenid" bson:"tokenid"`
}
