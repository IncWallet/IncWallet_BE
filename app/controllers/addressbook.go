package controllers

import (
	"errors"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/models"
)

/*
Address book controller
*/
type AddressBookCtrl struct {
	*revel.Controller
}

type AddressBookParam struct {
	Name           string `json:"name"`
	PaymentAddress string `json:"paymentaddress"`
	ChainName      string `json:"chainname"`
	ChainType      string `json:"chaintype"`
}

/*
Add Address
- name : hieutran
- paymentaddress:12Rtu5A4kF7QvtaXvrjreU36BKsLdjto8gr7bob8zmhReiJVgdEtyVNwkTqmJgNtP4sxhfxtDPAA5GDjCp6FtSaP9rg6Yn8ca1grNgH
- chainname: incognito
- chaintype: mainnet
*/
func (c AddressBookCtrl) AddAddress() revel.Result {
	addressParam := &AddressBookParam{}
	if err := c.Params.BindJSON(&addressParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}
	query := bson.M{
		"$or": []bson.M{
			bson.M{
				"name": addressParam.Name,
			},
			bson.M{
				"paymentaddress": addressParam.PaymentAddress,
			},
		},
	}
	if count ,err := database.AddressBook.Find(query).Count(); err == nil && count > 0 {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot add new address book"), errors.New(fmt.Sprintf("address {%v, %v} already existed", addressParam.Name, addressParam.PaymentAddress)), 0))
	}
	newAddress := &models.AddressBook{
		Name:           addressParam.Name,
		PaymentAddress: addressParam.PaymentAddress,
		ChainName:      addressParam.ChainName,
		ChainType:      addressParam.ChainType,
	}
	if err := database.AddressBook.Insert(newAddress); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot add new address book"), err.Error(), 0))
	}
	return c.RenderJSON(responseJsonBuilder(nil, "done", 0))
}

/*
Delete Address
- name
- paymentaddress
*/
func (c AddressBookCtrl) RemoveAddress() revel.Result {
	addressParam := &AddressBookParam{}
	if err := c.Params.BindJSON(&addressParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}

	addressBook := new(models.AddressBook)
	if err := database.AddressBook.Find(bson.M{
		"name": addressParam.Name,
		"paymentaddress": addressParam.PaymentAddress,
	}).One(&addressBook); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot remove address book"), errors.New(fmt.Sprintf("address {%v, %v} cannot be removed", addressParam.Name, addressParam.PaymentAddress)), 0))
	}

	if err := database.AddressBook.Remove(addressBook); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot remove address book"), err.Error(), 0))
	}
	return c.RenderJSON(responseJsonBuilder(nil, "done", 0))
}

/*
Get By Name
- name
*/
func (c AddressBookCtrl) GetByName() revel.Result {
	addressParam := &AddressBookParam{}
	if err := c.Params.BindJSON(&addressParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}

	addressBook := new(models.AddressBook)
	if err := database.AddressBook.Find(bson.M{
		"name": addressParam.Name,
	}).One(&addressBook); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get address book"), errors.New(fmt.Sprintf("address {%v} cannot be removed", addressParam.Name)), 0))
	}

	return c.RenderJSON(responseJsonBuilder(nil, addressBook, 0))
}

/*
Get By Payment Address
- name
*/
func (c AddressBookCtrl) GetByPaymentAddress() revel.Result {
	addressParam := &AddressBookParam{}
	if err := c.Params.BindJSON(&addressParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}

	addressBook := new(models.AddressBook)
	if err := database.AddressBook.Find(bson.M{
		"paymentaddress": addressParam.PaymentAddress,
	}).One(&addressBook); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get address book"), errors.New(fmt.Sprintf("address {%v} cannot be removed", addressParam.Name)), 0))
	}

	return c.RenderJSON(responseJsonBuilder(nil, addressBook, 0))
}

/*
Get All
*/
func (c AddressBookCtrl) GetAll() revel.Result {
	addressParam := &AddressBookParam{}
	if err := c.Params.BindJSON(&addressParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}
	addressBooks := make([]*models.AddressBook, 0)
	if err := database.AddressBook.Find(bson.M{}).All(&addressBooks); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get address book"), errors.New(fmt.Sprintf("address {%v} cannot be removed", addressParam.Name)), 0))
	}
	return c.RenderJSON(responseJsonBuilder(nil, addressBooks, 0))
}