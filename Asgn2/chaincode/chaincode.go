package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// smart contract
type Smart_Contract struct {
	contractapi.Contract
}

// account
// Key: ID, value: Balance
type Account struct {
	Balance int    `json:"Balance"`
	ID      string `json:"ID"` // {OrgMSP}
	Org     string `json:"Org"`
}

// item
// Key: ID, value: Name, Count, Price
type Item struct {
	Count int    `json:"Count"`
	ID    string `json:"ID"` // {OrgMSP}_{Name}
	Name  string `json:"Name"`
	Org   string `json:"Org"`
	Price int    `json:"Price"`
}

// initialize ledger
// add an account instance for calling Org with 0 balance
func (s *Smart_Contract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("InitLedger get client msp failed: %v", err)
	}
	account := Account{ID: OrgMSP, Org: OrgMSP[:len(OrgMSP)-3], Balance: 0}
	accountJSON, err := ctx.GetStub().GetState(account.ID)
	if err != nil {
		return fmt.Errorf("InitLedger get state account failed: %v", err)
	}
	if accountJSON != nil {
		return nil
	}
	accountJSON, err = json.Marshal(account)
	if err != nil {
		return fmt.Errorf("InitLedger json marshal account failed: %v", err)
	}
	err = ctx.GetStub().PutState(account.ID, accountJSON)
	if err != nil {
		return fmt.Errorf("InitLedger put state account failed: %v", err)
	}
	return nil
}

/// Marshal all Private data in JS app before passing to chaincode
func (s *Smart_Contract) AddBalance(ctx contractapi.TransactionContextInterface) error {
	type Input struct {
		Amount int `json:"amount"`
	}
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("<AddBalance> get client msp failed: %v", err)
	}
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("<AddBalance> get transient failed: %v", err)
	}
	amountJSON, exists := transient["amount"]
	if !exists {
		return fmt.Errorf("<AddBalance> amount not passed")
	}
	var input Input
	err = json.Unmarshal(amountJSON, &input)
	if err != nil {
		return fmt.Errorf("<AddBalance> unmarshal amount failed: %v", err)
	}
	if input.Amount < 0 {
		return fmt.Errorf("<AddBalance> cannot add negative amount")
	}
	accountJSON, err := ctx.GetStub().GetState(OrgMSP)
	if err != nil {
		return fmt.Errorf("<AddBalance> get state account failed: %v", err)
	}
	if accountJSON == nil {
		return fmt.Errorf("<AddBalance> account does not exist")
	}
	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return fmt.Errorf("<AddBalance> unmarshal account failed: %v", err)
	}
	account.Balance += input.Amount
	accountJSON, err = json.Marshal(account)
	if err != nil {
		return fmt.Errorf("<AddBalance> marshal account failed: %v", err)
	}
	err = ctx.GetStub().PutState(account.ID, accountJSON)
	if err != nil {
		return fmt.Errorf("<AddBalance> put state account failed: %v", err)
	}
	return nil
}

func (s *Smart_Contract) AddItem(ctx contractapi.TransactionContextInterface) error {
	type Input struct {
		Count int    `json:"count"`
		Name  string `json:"name"`
		Price int    `json:"price"`
	}
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("<AddItem> get client msp failed: %v", err)
	}
	collection := "_implicit_org_" + OrgMSP
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("<AddItem> get transient failed: %v", err)
	}
	itemJSON, exists := transient["item"]
	if !exists {
		return fmt.Errorf("<AddItem> item not passed")
	}
	var input Input
	err = json.Unmarshal(itemJSON, &input)
	if err != nil {
		return fmt.Errorf("<AddItem> unmarshal item failed: %v", err)
	}
	var item Item
	item.ID = OrgMSP + "_" + input.Name
	item.Org = OrgMSP[:len(OrgMSP)-3]
	item.Name = input.Name
	item.Count = input.Count
	item.Price = input.Price
	itemJSON, err = ctx.GetStub().GetPrivateData(collection, item.ID)
	if err != nil {
		return fmt.Errorf("<AddItem> get private data item failed: %v", err)
	}
	if itemJSON != nil {
		return fmt.Errorf("<AddItem> item already exists")
	}
	itemJSON, err = json.Marshal(item)
	if err != nil {
		return fmt.Errorf("<AddItem> marshal item failed: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(collection, item.ID, itemJSON)
	if err != nil {
		return fmt.Errorf("<AddItem> put private data item failed: %v", err)
	}
	return nil
}

func (s *Smart_Contract) GetBalance(ctx contractapi.TransactionContextInterface) (int, error) {
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return 0, fmt.Errorf("<GetBalance> get client identity failed: %v", err)
	}
	accountJSON, err := ctx.GetStub().GetState(OrgMSP)
	if err != nil {
		return 0, fmt.Errorf("<GetBalance> get state account failed: %v", err)
	}
	if accountJSON == nil {
		return 0, fmt.Errorf("<GetBalance> account does not exist")
	}
	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return 0, fmt.Errorf("<GetBalance> unmarshal account failed: %v", err)
	}
	return account.Balance, nil
}

func (s *Smart_Contract) GetItem(ctx contractapi.TransactionContextInterface) ([]Item, error) {
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return []Item{}, fmt.Errorf("<GetItem> get client msp failed: %v", err)
	}
	collection := "_implicit_org_" + OrgMSP
	var items []Item
	var item Item
	iter, err := ctx.GetStub().GetPrivateDataByRange(collection, "", "")
	if err != nil {
		return []Item{}, fmt.Errorf("<GetItem> get private data by range failed: %v", err)
	}
	for iter.HasNext() {
		entry, err := iter.Next()
		if err != nil {
			return []Item{}, fmt.Errorf("<GetItem> next iter failed: %v", err)
		}
		err = json.Unmarshal(entry.Value, &item)
		if err != nil {
			return []Item{}, fmt.Errorf("<GetItem> unmarshal item failed: %v", err)
		}
		items = append(items, item)
	}
	iter.Close()
	return items, nil
}

func (s *Smart_Contract) AddToMarket(ctx contractapi.TransactionContextInterface, name string, price int) error {
	if price < 0 {
		return fmt.Errorf("<AddToMarket> cannot have negative price")
	}
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("<AddToMarket> get client msp failed: %v", err)
	}
	id := OrgMSP + "_" + name
	collection := "_implicit_org_" + OrgMSP
	itemJSON, err := ctx.GetStub().GetPrivateData(collection, id)
	if err != nil {
		return fmt.Errorf("<AddToMarket> get private data item failed:%v", err)
	}
	if itemJSON == nil {
		return fmt.Errorf("<AddToMarket> item not found in inventory")
	}
	var item Item
	err = json.Unmarshal(itemJSON, &item)
	if err != nil {
		return fmt.Errorf("<AddToMarket> unmarshal item failed:%v", err)
	}
	item.Count--
	if item.Count == 0 {
		err = ctx.GetStub().DelPrivateData(collection, id)
		if err != nil {
			return fmt.Errorf("<AddToMarket> del private data item failed:%v", err)
		}
	} else {
		itemJSON, err = json.Marshal(item)
		if err != nil {
			return fmt.Errorf("<AddToMarket> marshal item failed:%v", err)
		}
		err = ctx.GetStub().PutPrivateData(collection, item.ID, itemJSON)
		if err != nil {
			return fmt.Errorf("<AddToMarket> put private data failed:%v", err)
		}
	}
	itemJSON, err = ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("<AddToMarket> get state failed:%v", err)
	}
	if itemJSON == nil {
		item.ID = OrgMSP + "_" + name
		item.Name = name
		item.Count = 1
		item.Price = price
		itemJSON, err = json.Marshal(item)
		if err != nil {
			return fmt.Errorf("<AddToMarket> marshal item failed:%v", err)
		}
		err = ctx.GetStub().PutState(item.ID, itemJSON)
		if err != nil {
			return fmt.Errorf("<AddToMarket> put state failed:%v", err)
		}
		err = ctx.GetStub().SetEvent("item_added", itemJSON)
		if err != nil {
			return fmt.Errorf("<AddToMarket> set event failed:%v", err)
		}
	} else {
		err = json.Unmarshal(itemJSON, &item)
		if err != nil {
			return fmt.Errorf("<AddToMarket> unmarshal item failed:%v", err)
		}
		item.Count++
		itemJSON, err = json.Marshal(item)
		if err != nil {
			return fmt.Errorf("<AddToMarket> marshal item failed:%v", err)
		}
		err = ctx.GetStub().PutState(item.ID, itemJSON)
		if err != nil {
			return fmt.Errorf("<AddToMarket> put state failed:%v", err)
		}
		err = ctx.GetStub().SetEvent("item_added", itemJSON)
		if err != nil {
			return fmt.Errorf("<AddToMarket> set event failed:%v", err)
		}
	}
	return nil
}

func (s *Smart_Contract) GetItemsInMarket(ctx contractapi.TransactionContextInterface) ([]Item, error) {
	var items []Item
	var item Item
	iter, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return []Item{}, fmt.Errorf("<GetItemsInMarket> get state by range failed: %v", err)
	}
	for iter.HasNext() {
		entry, err := iter.Next()
		if err != nil {
			return []Item{}, fmt.Errorf("<GetItemsInMarket> Iter Next Failed: %v", err)
		}
		if strings.Contains(entry.Key, "_") {
			err = json.Unmarshal(entry.Value, &item)
			if err != nil {
				return []Item{}, fmt.Errorf("<GetItemsInMarket> unmarshal item failed: %v", err)
			}
			items = append(items, item)
		}
	}
	iter.Close()
	return items, nil
}

func (s *Smart_Contract) BuyFromMarket(ctx contractapi.TransactionContextInterface, id string) error {
	OrgMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> get msp id failed: %v", err)
	}
	itemJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> get state item failed: %v", err)
	}
	if itemJSON == nil {
		return fmt.Errorf("item not available in market")
	}
	var item Item
	err = json.Unmarshal(itemJSON, &item)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> unmarshal item failed: %v", err)
	}
	accountJSON, err := ctx.GetStub().GetState(OrgMSP)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> get state account failed: %v", err)
	}
	if accountJSON == nil {
		return fmt.Errorf("account does not exist")
	}
	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> unmarshal account failed: %v", err)
	}
	if account.Balance < item.Price {
		return fmt.Errorf("insufficient balance in account")
	}
	account.Balance -= item.Price
	accountJSON, err = json.Marshal(account)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> marshal account failed: %v", err)
	}
	err = ctx.GetStub().PutState(account.ID, accountJSON)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> put state account failed: %v", err)
	}
	item.Count--
	if item.Count == 0 {
		err = ctx.GetStub().DelState(item.ID)
		if err != nil {
			return fmt.Errorf("<BuyFromMarket> del state item failed: %v", err)
		}
	} else {
		itemJSON, err = json.Marshal(item)
		if err != nil {
			return fmt.Errorf("<BuyFromMarket> marshal item failed: %v", err)
		}
		err = ctx.GetStub().PutState(item.ID, itemJSON)
		if err != nil {
			return fmt.Errorf("<BuyFromMarket> put state item failed: %v", err)
		}
	}
	OrgMSP = id[:7]
	accountJSON, err = ctx.GetStub().GetState(OrgMSP)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> get state account failed: %v", err)
	}
	if accountJSON == nil {
		return fmt.Errorf("<BuyFromMarket> account does not exist: %v", err)
	}
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> unmarshal account failed: %v", err)
	}
	account.Balance += item.Price
	accountJSON, err = json.Marshal(account)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> marshal account failed: %v", err)
	}
	err = ctx.GetStub().PutState(account.ID, accountJSON)
	if err != nil {
		return fmt.Errorf("<BuyFromMarket> put state account failed: %v", err)
	}
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(Smart_Contract))
	if err != nil {
		fmt.Printf("error creating chaincode")
		return
	}
	err = chaincode.Start()
	if err != nil {
		fmt.Printf("error starting chaincode")
		return
	}
}
