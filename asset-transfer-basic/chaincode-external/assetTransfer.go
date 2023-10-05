/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Make serverConfig struct
type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing a transaction
type SmartContract struct {
	contractapi.Contract
}

// TransactionPurchase struct for transaction of type purchase, PUREPurchase
// Each TransactionPurchase has a Name, AmountInGrams, Cost, Category, FedTax, WeFee, StateTax, AmountInMiligramsOfTotalTHC
// Write in alphabetical order for Marshal
type TransactionPurchase struct {
	AmountInGrams               float64 `json:"AmountInGrams"`
	AmountInMiligramsOfTotalTHC float64 `json:"AmountInMiligramsOfTotalTHC"`
	Category                    string  `json:"Category"`
	Cost                        float64 `json:"Cost"`
	FedTax                      float64 `json:"FedTax"`
	Name                        string  `json:"Name"`
	StateTax                    float64 `json:"StateTax"`
	THCPercent                  float64 `json:"THCPercent"`
	WeFee                       float64 `json:"WeFee"`
}

// Asset describes basic details of what makes up a simple transaction
// Each transaction has a unique ID, Amount, location, name of sender, name of reciever, timestamp, typeOfTransaction: TransactionPurchase or PUREUP or PUREEX or PUREWI, status
// Write in alphabetical order for Marshal
type Transaction struct {
	Amount              float64             `json:"Amount"`
	Location            string              `json:"Location"`
	Reciever            string              `json:"Reciever"`
	Sender              string              `json:"Sender"`
	Status              string              `json:"Status"`
	Timestamp           string              `json:"Timestamp"`
	TransactionID       string              `json:"TransactionID"`
	TransactionPurchase TransactionPurchase `json:"TransactionPurchase"`
	TypeOfTransaction   string              `json:"TypeOfTransaction"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Transaction
}

// InitLedger adds a base set of transactions to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	transactions := []Transaction{
		{Amount: 100, Location: "40.6935385, -73.8555598", Reciever: "New York Dispensary 1", Sender: "Mary Anne Kate", Status: "Pending", Timestamp: "2023-09-14 18:00:23", TransactionID: "1", TransactionPurchase: TransactionPurchase{AmountInGrams: 14, AmountInMiligramsOfTotalTHC: 2800, Category: "Flower", Cost: 100, FedTax: 10, Name: "Green Crack", StateTax: 14, THCPercent: .20, WeFee: 1}, TypeOfTransaction: "PUREPU"},
		{Amount: 50, Location: "40.6935385, -73.8555598", Reciever: "New York Dispensary 2", Sender: "John Jones", Status: "Pending", Timestamp: "2023-09-14 19:00:23", TransactionID: "2", TransactionPurchase: TransactionPurchase{AmountInGrams: 14, AmountInMiligramsOfTotalTHC: 2800, Category: "Flower", Cost: 50, FedTax: 5, Name: "Sour Diesel", StateTax: 10, THCPercent: .20, WeFee: 1}, TypeOfTransaction: "PUREPU"},
		{Amount: 75, Location: "45.6935385, -69.8555598", Reciever: "New York Dispensary 3", Sender: "Jimmy Nixon", Status: "Pending", Timestamp: "2023-09-17 20:00:23", TransactionID: "3", TransactionPurchase: TransactionPurchase{AmountInGrams: 14, AmountInMiligramsOfTotalTHC: 2800, Category: "Flower", Cost: 75, FedTax: 7, Name: "Blue Dream", StateTax: 9, THCPercent: .20, WeFee: 1}, TypeOfTransaction: "PUREPU"},
		{Amount: 100, Location: "40.6935385, -73.8555598", Reciever: "New York Dispensary 4", Sender: "Jimmy Cricket", Status: "Pending", Timestamp: "2023-09-19 18:00:23", TransactionID: "4", TransactionPurchase: TransactionPurchase{AmountInGrams: 14, AmountInMiligramsOfTotalTHC: 2800, Category: "Flower", Cost: 100, FedTax: 10, Name: "Northern Lights", StateTax: 14, THCPercent: .20, WeFee: 1}, TypeOfTransaction: "PUREPU"},
	}
	for _, transaction := range transactions {
		transactionJSON, err := json.Marshal(transaction)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(transaction.TransactionID, transactionJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil
}

//CreateTransaction adds a new transaction to the world state with given details

func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, transactionID string, amount float64, location string, reciever string, sender string, status string, timestamp string, transactionPurchase TransactionPurchase, typeOfTransaction string) error {
	exists, err := s.TransactionExists(ctx, transactionID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the transaction %s already exists", transactionID)
	}
	transaction := Transaction{
		Amount:              amount,
		Location:            location,
		Reciever:            reciever,
		Sender:              sender,
		Status:              status,
		Timestamp:           timestamp,
		TransactionID:       transactionID,
		TransactionPurchase: transactionPurchase,
		TypeOfTransaction:   typeOfTransaction,
	}
	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(transactionID, transactionJSON)
}

// ReadTransaction returns the transaction stored in the world state with given id
func (s *SmartContract) ReadTransaction(ctx contractapi.TransactionContextInterface, transactionID string) (*Transaction, error) {
	transactionJSON, err := ctx.GetStub().GetState(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if transactionJSON == nil {
		return nil, fmt.Errorf("the transaction %s does not exist", transactionID)
	}
	var transaction Transaction
	err = json.Unmarshal(transactionJSON, &transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// ReadTransaction from the world state with given location
func (s *SmartContract) ReadTransactionByLocation(ctx contractapi.TransactionContextInterface, location string) ([]*Transaction, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all transactions in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("location~name", []string{location})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var transactions []*Transaction
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}
	return transactions, nil
}

//UpdateTransaction updates an existing transaction in the world state with provided parameters

func (s *SmartContract) UpdateTransaction(ctx contractapi.TransactionContextInterface, transactionID string, amount float64, location string, reciever string, sender string, status string, timestamp string, transactionPurchase TransactionPurchase, typeOfTransaction string) error {
	exists, err := s.TransactionExists(ctx, transactionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the transaction %s does not exist", transactionID)
	}
	// overwriting original transaction with new transaction
	transaction := Transaction{
		Amount:              amount,
		Location:            location,
		Reciever:            reciever,
		Sender:              sender,
		Status:              status,
		Timestamp:           timestamp,
		TransactionID:       transactionID,
		TransactionPurchase: transactionPurchase,
		TypeOfTransaction:   typeOfTransaction,
	}
	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(transactionID, transactionJSON)
}

// DeleteTransaction deletes an given transaction from the world state
func (s *SmartContract) DeleteTransaction(ctx contractapi.TransactionContextInterface, transactionID string) error {
	exists, err := s.TransactionExists(ctx, transactionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the transaction %s does not exist", transactionID)
	}
	return ctx.GetStub().DelState(transactionID)
}

// TransactionExists returns true when transaction with given ID exists in world state
func (s *SmartContract) TransactionExists(ctx contractapi.TransactionContextInterface, transactionID string) (bool, error) {
	transactionJSON, err := ctx.GetStub().GetState(transactionID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return transactionJSON != nil, nil
}

// GetAllTransactions returns all transactions found in world state
func (s *SmartContract) GetAllTransactions(ctx contractapi.TransactionContextInterface) ([]*Transaction, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all transactions in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var transactions []*Transaction
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &transaction)
	}
	return transactions, nil
}

func main() {

	//Read config from env variables- chaincode.env

	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	transactionLedgerChaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("Error creating transactionLedger chaincode: %v", err)
	}

	//Create chaincode server
	server := &shim.ChaincodeServer{
		CCID:     config.CCID,
		Address:  config.Address,
		CC:       transactionLedgerChaincode,
		TLSProps: getTLSProperties(),
	}

	if err := server.Start(); err != nil {
		log.Panicf("Error starting transactionLedger chaincode: %v", err)
	}
}

// getTLSProperties returns TLS properties for chaincode server
func getTLSProperties() shim.TLSProperties {
	//Check if chaincode is TLS enabled
	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")
	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")
	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")
	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	//convert tlsDisabledStr to bool
	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)
	var keyBytes, certBytes, clientCACertBytes []byte
	var err error

	if !tlsDisabled {
		keyBytes, err = os.ReadFile(key)
		if err != nil {
			log.Panicf("Failed to load key, error while reading the crypto file: %v", err)
		}
		certBytes, err = os.ReadFile(cert)
		if err != nil {
			log.Panicf("Failed to load cert, error while reading the crypto file: %v", err)
		}
	}

	//Did not request for the peer cert verification
	if clientCACert != "" {
		clientCACertBytes, err = os.ReadFile(clientCACert)
		if err != nil {
			log.Panicf("Failed to load clientCACert, error while reading the crypto file: %v", err)
		}
	}

	return shim.TLSProperties{
		Disabled:      tlsDisabled,
		Key:           keyBytes,
		Cert:          certBytes,
		ClientCACerts: clientCACertBytes,
	}
}

func getEnvOrDefault(env, defaultVal string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		return defaultVal
	}
	return value
}

//Note that the method returns default value if the string cannot be parsed

func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}
