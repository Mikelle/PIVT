///////////////////////////////////////////////////////////////
// File containing the model structs for the Cache Coin Token,
// and blockchain inputs and outputs
///////////////////////////////////////////////////////////////

package main

/*---------------------------------------------------------------------------
SMART CONTRACT MODELS FOR LIQUIDITY POOLS
-----------------------------------------------------------------------------*/

// Definition of the model of the smart contract //
type DataProtocolSmartContract struct{}

// // Definition the output for the smart contract //
// type Output struct {
// 	ID                	    string    					`json:"ID"`
// 	DID                	    string    					`json:"DID"`
// 	UpdateWallets           map[string]MultiWallet      `json:"UpdateWallets"`
// 	UpdateUsers             map[string]Actor            `json:"UpdateUsers"`
// }

// Definition of an actor in the Cache Ecosystem //
type Actor struct {
	PublicId      string          `json:"PublicId"`
	PublicAddress string          `json:"PublicAddress"`
	Role          string          `json:"Role"`
	Privacy       map[string]bool `json:"Privacy"`
}

// // Definition of the encryption object with DIDs //
// type Encryption struct {
// 	PublicId        string `json:"PublicId"`
// 	DID             string `json:"DID"`
// }

// // Definition of the privacy modifier object //
// type PrivacyModifier struct {
// 	PublicId        string `json:"PublicId"`
// 	BusinessId      string `json:"BusinessId"`
// 	Enabled         bool   `json:"Enabled"`
// }

// // Definition of the insigth discovery object //
// type InsightDiscovery struct {
// 	Business_Id     string   `json:"Business_Id"`
// 	DID_list        []string `json:"DID_list"`
// 	ID_list         []string `json:"ID_list"`
// }

// // Definition of the insigth purchase object //
// type InsightPurchase struct {
// 	DID_list        		[]string 				`json:"DID_list"`
// 	Price           		float64  				`json:"Price"`
// 	Business_Id     		string   				`json:"Business_Id"`
// 	InsightDistribution     map[string]float64 		`json:"InsightDistribution"`
// }

// // Definition of the insigth target object //
// type InsightTarget struct {
// 	TID_list        		[]string 				`json:"TID_list"`
// 	Business_Id     		string   				`json:"Business_Id"`
// }

/*---------------------------------------------------------------------------
COIN BALANCE INVOKATIONS
-----------------------------------------------------------------------------*/
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
	Date           float64 `json:"Date"`
}

/*---------------------------------------------------------------------------
-----------------------------------------------------------------------------*/
