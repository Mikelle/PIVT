package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func (obj *Token) ToLedgerValue() ([]byte, error) {
	return json.Marshal(obj)
}

func (obj *Token) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	attributes := []string{obj.Symbol}

	return stub.CreateCompositeKey(IndexToken, attributes)
}

func (obj *Token) SaveState(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := obj.ToCompositeKey(stub)
	if err != nil {
		message := fmt.Sprintf("unable to create a composite key: %s", err.Error())
		return errors.New(message)
	}
	var ledgerValue []byte
	ledgerValue, err = obj.ToLedgerValue()
	if err != nil {
		message := fmt.Sprintf("unable to compose a ledger value: %s", err.Error())
		return errors.New(message)
	}

	return stub.PutState(compositeKey, ledgerValue)
}

// returns false if an Account object wasn't found in the ledger; otherwise returns true
func (obj *Token) LoadState(stub shim.ChaincodeStubInterface) (bool, error) {
	compositeKey, err := obj.ToCompositeKey(stub)
	if err != nil {
		message := fmt.Sprintf("unable to create a composite key: %s", err.Error())
		return false, errors.New(message)
	}

	var ledgerValue []byte
	ledgerValue, err = stub.GetState(compositeKey)
	if err != nil {
		message := fmt.Sprintf("unable to read the ledger value: %s", err.Error())
		return false, errors.New(message)
	}

	if ledgerValue == nil {
		return false, nil
	}

	return true, json.Unmarshal(ledgerValue, &obj)
}
