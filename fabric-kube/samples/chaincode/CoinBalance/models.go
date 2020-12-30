///////////////////////////////////////////////////////////////
// File containing the model structs for the Cache Coin Token,
// and blockchain inputs and outputs
///////////////////////////////////////////////////////////////

package main

/*---------------------------------------------------------------------------
SMART CONTRACT MODELS
-----------------------------------------------------------------------------*/

// Definition of the model of the smart contract //
type CoinBalanceSmartContract struct{}

// Definition the output for the smart contract //
type Output struct {
	UpdateBalances map[string]Balance  `json:"UpdateBalances"`
	UpdateTokens   map[string]Token    `json:"UpdateTokens"`
	Transactions   map[string]Transfer `json:"Transactions"`
}

// Definition of the user Balance for a given token //
type Balance struct {
	Address    string  `json:"Address"`
	Token      string  `json:"Token"`
	Amount     float64 `json:"Amount"`
	Credit     float64 `json:"Credit"`
	LockUpDate int64   `json:"LockUpDate"`
}

// Definition of the user Balance for a given token //
type FinancialScores struct {
	TrustScore       float64 `json:"TrustScore"`
	EndorsementScore float64 `json:"EndorsementScore"`
}

// Definition of a Token Transfer //
type Transfer struct {
	Type           string  `json:"Type"`
	Token          string  `json:"Token"`
	From           string  `json:"From"`
	To             string  `json:"To"`
	AvoidCheckTo   bool    `json:"AvoidCheckTo"`
	AvoidCheckFrom bool    `json:"AvoidCheckFrom"`
	Amount         float64 `json:"Amount"`
	Id             string  `json:"Id"`
	Date           int64   `json:"Date"`
}

// Definition of Token Objects in Blockchain //
type Token struct {
	Name       string  `json:"Name"`
	TokenType  string  `json:"TokenType"`
	Symbol     string  `json:"Symbol"`
	Supply     float64 `json:"Supply"`
	LockUpDate int64   `json:"LockUpDate"`
}

// Definition of a Token Swapping //
type MultiMinter struct {
	Token       string             `json:"Token"`
	Type        string             `json:"Type"`
	TxnId       string             `json:"TxnId"`
	FromAddress string             `json:"FromAddress"`
	Date        string             `json:"Date"`
	TotalAmount float64            `json:"TotalAmount"`
	Transfers   map[string]float64 `json:"Transfers"`
}

// Definition of a user history retrieval //
type History struct {
	PublicId  string `json:"PublicId"`
	Timestamp int64  `json:"Timestamp"`
}

// Definition of a Token Swapping //
type Swapper struct {
	PublicId string  `json:"PublicId"`
	Token    string  `json:"Token"`
	Amount   float64 `json:"Amount"`
	TxnId    string  `json:"TxnId"`
	Date     int64   `json:"Date"`
}

/*---------------------------------------------------------------------------
-----------------------------------------------------------------------------*/
