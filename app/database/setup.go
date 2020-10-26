package database

import "gopkg.in/mgo.v2"

/*
Database session
*/
var Session *mgo.Session

/*
Book's models connection
*/
var Wallet *mgo.Collection
var Accounts *mgo.Collection
var Coins *mgo.Collection
var State *mgo.Collection
var TxHistory *mgo.Collection
var Tokens *mgo.Collection
var PoolPairs *mgo.Collection
var PdeHistory *mgo.Collection
var AddressBook *mgo.Collection
var Committee *mgo.Collection
var Reward *mgo.Collection
/*
Init database
*/
func Init(uri, dbname string) error {
	session, err := mgo.Dial(uri)
	if err != nil {
		return err
	}

	// See https://godoc.org/labix.org/v2/mgo#Session.SetMode
	session.SetMode(mgo.Monotonic, true)

	// Expose session and models
	Session = session
	Wallet = session.DB(dbname).C("wallet")
	Accounts = session.DB(dbname).C("accounts")
	Coins = session.DB(dbname).C("coins")
	State = session.DB(dbname).C("state")
	TxHistory = session.DB(dbname).C("txhistory")
	Tokens = session.DB(dbname).C("tokens")
	PoolPairs = session.DB(dbname).C("poolpairs")
	PdeHistory = session.DB(dbname).C("pdehistory")
	AddressBook = session.DB(dbname).C("addressbook")
	Committee = session.DB(dbname).C("committee")
	Reward = session.DB(dbname).C("reward")
	return nil
}
