package models

type Tokens struct {
	ID                 string  `json:"id" bson:"id"`
	Symbol             string  `json:"symbol" bson:"symbol"`
	Name               string  `json:"name" bson:"name"`
	Decimal           int     `json:"decimal" bson:"decimal"`
	EDecimal          int     `json:"edecimals" bson:"edecimal"`
	ESymbol            string  `json:"esymbol" bson:"esymbol"`
	Verified           bool    `json:"verified" bson:"verified"`
	Network            string  `json:"network" bson:"network"`
	Image              string  `json:"image" bson:"image"`
}
