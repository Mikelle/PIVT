package main

import (
	//"time"
	"encoding/json"
	"fmt"
	"math"

	//"strings"
	//"bytes"
	//"math"
	//"strconv"
	//"github.com/rs/xid"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	//"github.com/hyperledger/fabric/common/util"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/* -------------------------------------------------------------------------------------------------
Init:  this function is called at PRIVI Blockchain Deployment and initialises the Coin Balance
	   Smart Contract. This smart contract is the responsible to manage the balances of the
	   different tokens powered in PRIVI Ecosystem. The initialisation sets the Private Key ID
	   of the admin in the ledger. Args: array containing a string:
PrivateKeyID           string   // Private Key of the admin of the smart contract
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()
	if args[0] == "UPGRADE" {
		return shim.Success(nil)
	}

	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
Invoke:  this function is the router of the different functions supported by the Smart Contract.
		 It receives the input from the controllers and ensure the correct calling of the functions.
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// Retrieve function and arguments //
	function, args := stub.GetFunctionAndParameters()

	// Call the proper function //
	switch function {

	case "registerToken":
		err := checkPermissions(stub, ADMIN_ROLE, function)
		if err != nil {
			return shim.Error(err.Error())
		}
		return t.registerToken(stub, args)

	case "removeToken":
		err := checkPermissions(stub, ADMIN_ROLE, function)
		if err != nil {
			return shim.Error(err.Error())
		}
		return t.removeToken(stub, args)

	case "getTokenInfoByType":
		tokenList, err := t.getTokenInfoByType(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		tokenListBytes, _ := json.Marshal(tokenList)
		return shim.Success(tokenListBytes)

	case "getTokenListByType":
		tokenList, err := t.getTokenListByType(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		tokenListBytes, _ := json.Marshal(tokenList)
		return shim.Success(tokenListBytes)

	case "getToken":
		token, err := t.getToken(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		tokenBytes, _ := json.Marshal(token)
		return shim.Success(tokenBytes)

	case "checkAddressExist":
		exist := t.checkAddressExist(stub, args[0])
		if !exist {
			return shim.Error("ERROR: ADDRESS DOES NOT EXISTS")
		}
		return shim.Success(nil)

	case "getWalletType":
		return t.getWalletType(stub, args)

	case "balanceOf":
		return t.balanceOf(stub, args)

	case "mint":
		err := checkPermissions(stub, ADMIN_ROLE, function)
		if err != nil {
			return shim.Error(err.Error())
		}
		return t.mint(stub, args)

	case "burn":
		return t.burn(stub, args)

	case "transfer":
		return t.transfer(stub, args)

	case "multitransfer":
		return t.multitransfer(stub, args)

	case "initialiseBalance":
		return t.initialiseBalance(stub, args)

	case "initialiseFinancialScores":
		return t.initialiseFinancialScores(stub, args)

	case "updateFinancialScores":
		return t.updateFinancialScores(stub, args)

	case "getFinancialScores":
		return t.getFinancialScores(stub, args)

	case "getBalancesOfAddress":
		return t.getBalancesOfAddress(stub, args)

	case "getBalancesOfTokenHolders":
		return t.getBalancesOfTokenHolders(stub, args)

	case "getTokenHolderList":
		holderList, err := getTokenHolderList(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		holderListBytes, _ := json.Marshal(holderList)
		return shim.Success(holderListBytes)

	case "updateTokenInfo":
		err := checkPermissions(stub, ADMIN_ROLE, function)
		if err != nil {
			return shim.Error(err.Error())
		}
		return t.updateTokenInfo(stub, args)

	}

	// If function does not exist, retrieve error//
	return shim.Error("ERROR: INCORRECT FUNCTION NAME " +
		function)
}

/* -------------------------------------------------------------------------------------------------
registerToken:  this function is called when a new token is listed on the PRIVI Blockchain. It can
				only be called by Admin. At deployment, PRIVI Coin and Base Coin are created in the
				system. Input: array containing a json with fields:
Name           string   // Name of the Token  (args[0])
TokenType      string   // Type of Token to register
TokenSymbol    string   // Symbol of the Token
Supply         string   // Supply introduced at creation
LockUpDate     string   // If the token has a lock up date which prevent to be transfered
Address        string   // Address to input initial supply (args[1])
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) registerToken(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve arguments of the input in a Token Model struct //
	if len(args) != 2 {
		return shim.Error("ERROR: REGISTER TOKEN FUNCTION SHOULD BE CALLED " +
			"WITH 2 ARGUMENTS.")
	}
	token := Token{}
	json.Unmarshal([]byte(args[0]), &token)

	// Check if Token is already registered on Blockchain //
	tokenCheck, err := checkTokenRegistered(stub, token.Symbol)
	if err != nil {
		return shim.Error(err.Error())
	}
	if tokenCheck {
		return shim.Error("ERROR: TOKEN " + token.Symbol + " ALREADY REGISTERED " +
			"ON THE SYSTEM. ")
	}

	// Check if integer condition if we have an NFT Pod Token //
	if token.TokenType == NFT_POD_TOKEN {
		if token.Supply != math.Trunc(token.Supply) {
			return shim.Error("ERROR: THE SUPPLY OF NFT POD TOKEN " +
				" SHOULD BE AN INTEGER.")
		}
	}

	// Add new token on Blockchain //
	err = t.updateToken(stub, token)
	if err != nil {
		return shim.Error(err.Error())
	}
	tokens := make(map[string]Token)
	tokens[token.Symbol] = token

	// Update address with the initial supply //
	balance := Balance{
		Token: token.Symbol, Address: args[1],
		Amount: token.Supply, Credit: 0.}
	err = t.updateBalance(stub, balance)
	balances := make(map[string]Balance)
	balances[args[1]+" "+token.Symbol] = balance

	return generateOutput(balances, tokens, nil)

}

/* -------------------------------------------------------------------------------------------------
removeToken:  this function is called when a pod Token is removed from the system.
Symbol         string   // Symbol of the Token
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) removeToken(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	token, err := t.getToken(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(IndexToken+token.Symbol, nil)
	if err != nil {
		return shim.Error("ERROR: DELETING TOKEN " + token.Symbol +
			" ON BLOCKCHAIN. " + err.Error())
	}
	return shim.Success(nil)

}

/* -------------------------------------------------------------------------------------------------
balanceOf: This function is called to get the balance on the wallet of a given actor. The caller of
		   this function can only be Cache Admin or wallet owner.
		   Args: array containing
args[0]               string   // Public Id Key of the actor
args[1]               string   // Token Symbol

// digestHash -> string
// signature -> string
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) balanceOf(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve information from the input //
	if len(args) != 2 {
		return shim.Error("ERROR: BALANCEOF FUNCTION SHOULD BE CALLED " +
			"WITH TWO ARGUMENT.")
	}

	// Get balance //
	balance, err := t.checkBalance(stub, args[0], args[1], true)
	if err != nil {
		return shim.Error(err.Error())
	}
	balanceBytes, _ := json.Marshal(balance)
	return shim.Success(balanceBytes)
}

/* -------------------------------------------------------------------------------------------------
intialiseBalance: this function initialises a new balance for a user given a token
PublicId                string    // Id of the address
Token                   float64   // Token to initialise balance
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) initialiseBalance(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve information from the input //
	if len(args) != 2 {
		return shim.Error("ERROR: INITIALISEBALANCE FUNCTION SHOULD BE CALLED " +
			"WITH TWO ARGUMENTS.")
	}

	// Check if balance exists. Otherwise it creates zero balance //
	balance, err := t.checkBalance(stub, args[0], args[1], true)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Store balance on Blockchain //
	err = t.updateBalance(stub, balance)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
registerWallet: this function initialises a wallet for a pool
Id                string    // Id of the address
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) registerAddress(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Check that wallet is not already registered //
	addressExist := t.checkAddressExist(stub, args[0])
	if addressExist {
		return shim.Error("ERROR: ADDRESS " + args[0] +
			" IS ALREADY REGISTERED IN THE SYSTEM.")
	}

	// Register new address on blockchain //
	input, _ := json.Marshal([]byte(args[0]))
	err := stub.PutState(IndexWallets+args[0], input)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
initialiseFinancialScores: this function updates the Trust and Endorsement scores of an user
publicId                string    // Id of the user  (args[0])
TrustScore              float64   // Public Id Key of the actor
EndorsementScore        float64   // Token Symbol (args[1])
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) initialiseFinancialScores(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve information from the input //
	if len(args) != 2 {
		return shim.Error("ERROR: INITIALISEFINANCIALSCORES FUNCTION SHOULD BE CALLED " +
			"WITH TWO ARGUMENTS.")
	}
	scores := FinancialScores{}
	json.Unmarshal([]byte(args[1]), &scores)

	// Check correctness of scores //
	inRange := checkRange(scores.TrustScore, 0., 1.)
	if !inRange {
		return shim.Error("ERROR: TRUST SCORE SHOULD BE BETWEEN 0 AND 1")
	}
	inRange = checkRange(scores.EndorsementScore, 0., 1.)
	if !inRange {
		return shim.Error("ERROR: ENDORSEMENT SCORE SHOULD BE BETWEEN 0 AND 1")
	}

	// Update user scores on Blockchain //
	scoresBytes, _ := json.Marshal(scores)
	err := stub.PutState(IndexFinancialScores+args[0], scoresBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
updateFinancialScores: this function updates the Trust and Endorsement scores of an user
publicId                string    // Id of the user  (args[0])
TrustScore              float64   // Public Id Key of the actor
EndorsementScore        float64   // Token Symbol (args[1])
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) updateFinancialScores(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve information from the input //
	if len(args) != 2 {
		return shim.Error("ERROR: UPDATEFINANCIALSCORES FUNCTION SHOULD BE CALLED " +
			"WITH TWO ARGUMENTS.")
	}
	scores := FinancialScores{}
	json.Unmarshal([]byte(args[1]), &scores)

	// Check that user exists //
	err := t.checkUserExist(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	// Check correctness of scores //
	inRange := checkRange(scores.TrustScore, 0., 1.)
	if !inRange {
		return shim.Error("ERROR: TRUST SCORE SHOULD BE BETWEEN 0 AND 1")
	}
	inRange = checkRange(scores.EndorsementScore, 0., 1.)
	if !inRange {
		return shim.Error("ERROR: ENDORSEMENT SCORE SHOULD BE BETWEEN 0 AND 1")
	}

	// Update user scores on Blockchain //
	scoresBytes, _ := json.Marshal(scores)
	err = stub.PutState(IndexFinancialScores+args[0], scoresBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
getFinancialScores: this function retrieves the financial scores of an user
publicId                string    // Id of the user  (args[0])
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getFinancialScores(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	scoresBytes, err := stub.GetState(IndexFinancialScores + args[0])
	if err != nil {
		return shim.Error("ERROR: GETTING THE FINANCIAL SCORES OF THE USER")
	}
	if scoresBytes == nil {
		return shim.Error("ERROR: THE USER HAS NOT REGISTER FINANCIAL SCORES")
	}
	return shim.Success(scoresBytes)
}

/* -------------------------------------------------------------------------------------------------
getWalletType: this function retrieves the wallet of an user for a given type of token
publicId                string    // Id of the user
Type                    string    // Type of tokens to retrieve
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getWalletType(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Get Token List //
	tokenList, err := t.getTokenListByType(stub, args[1])
	if err != nil {
		return shim.Error(err.Error())
	}

	// Retrieve all the balances of user //
	balances := make(map[string]Balance)
	for _, token := range tokenList {
		balances[token], err = t.checkBalance(stub, args[0], token, true)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// Prepare Output //
	balancesBytes, _ := json.Marshal(balances)
	return shim.Success(balancesBytes)
}

/* -------------------------------------------------------------------------------------------------
transfer: This function is called to transfer a given token from one wallet to another one.
Token              string   // Symbol of token to transfer
From               string   // Id of the sender
To                 string   // Id of the receiver
Amount             float64  // Amount that is being sent
Id                 string   // ID of the transaction
Date               float64  // Date timestamp
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) transfer(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieve information from the input in a Transfer object //
	// if len(args) != 1 {
	// 	return shim.Error("ERROR: TRANSFER FUNCTION SHOULD BE CALLED " +
	// 		"WITH ONE ARGUMENT.")
	// }

	transfer := Transfer{}
	json.Unmarshal([]byte(args[0]), &transfer)

	// Validate From transaction //
	err := validateSignature(transfer.From, args[1], args[2])
	if err != nil {
		return shim.Error(err.Error())
	}

	transactions := make(map[string]Transfer)
	balances := make(map[string]Balance)
	var check bool

	// Check if transfer is possible //
	err = t.checkTokenTransferConditions(stub, transfer.Token,
		transfer.Date, transfer.Amount)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Verify the sender and receiver are not the same //
	if transfer.From == transfer.To {
		return shim.Error("ERROR: SENDER AND RECEIVER CANNOT BE THE " +
			" SAME IN A TRANSFER.")
	}

	// Retrieve balance of sender and receiver //
	check = true
	if transfer.AvoidCheckFrom {
		check = false
	}
	senderBalance, err1 := t.checkBalance(stub, transfer.From, transfer.Token,
		check)
	if err1 != nil {
		return shim.Error(err1.Error())
	}

	check = true
	if transfer.AvoidCheckTo {
		check = false
	}
	receiverBalance, err2 := t.checkBalance(stub, transfer.To, transfer.Token,
		check)
	if err2 != nil {
		return shim.Error(err2.Error())
	}

	// Transfer funds  //
	senderBalance.Amount, receiverBalance.Amount, err = t.transferHelper(
		stub, senderBalance.Amount, receiverBalance.Amount, transfer.Amount)
	if err != nil {
		return shim.Error(err.Error())
	}
	transfer.Type = "Transfer"
	transactions[transfer.Id] = transfer

	// Update balances of sender and receiver //
	err = t.updateBalance(stub, senderBalance)
	if err != nil {
		return shim.Error(err.Error())
	}
	balances[transfer.From+" "+transfer.Token] = senderBalance

	err = t.updateBalance(stub, receiverBalance)
	if err != nil {
		return shim.Error(err.Error())
	}
	balances[transfer.From+" "+transfer.Token] = receiverBalance

	// Prepare output object with updates //
	return generateOutput(balances, nil, transactions)
}

/* -------------------------------------------------------------------------------------------------
multitransfer: This function is called to perform a multitransfer between different actors in
               one call to blockchain. Args: is a list of transfer types.
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) multitransfer(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Set of users participating in transfer and whose state should be updated //
	transactions := make(map[string]Transfer)
	balances := make(map[string]Balance)
	var err error
	var senderBalance Balance
	var receiverBalance Balance
	var inList bool
	var check bool

	// Iterate throught all the transactions on the multitransfer call //
	for _, arg := range args {

		// Retrieve trasnfer from the list //
		transfer := Transfer{}
		json.Unmarshal([]byte(arg), &transfer)
		if transfer.From == transfer.To || transfer.Amount == 0 {
			continue
		}

		// Check if transfer is allowed //
		err = t.checkTokenTransferConditions(stub, transfer.Token,
			transfer.Date, transfer.Amount)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Check if sender is already in transaction users list //
		senderBalance, inList = balances[transfer.From+" "+transfer.Token]
		if !inList {
			check = true
			if transfer.AvoidCheckFrom {
				check = false
			}
			senderBalance, err = t.checkBalance(stub, transfer.From, transfer.Token, check)
			if err != nil {
				return shim.Error(err.Error())
			}
		}

		// Check if receiver is already in transaction users list //
		receiverBalance, inList = balances[transfer.To+" "+transfer.Token]
		if !inList {
			check = true
			if transfer.AvoidCheckTo {
				check = false
			}
			receiverBalance, err = t.checkBalance(stub, transfer.To, transfer.Token, check)
			if err != nil {
				return shim.Error(err.Error())
			}
		}

		// Check that the sender holds the amount to send and transfer funds //
		senderBalance.Amount, receiverBalance.Amount, err = t.transferHelper(
			stub, senderBalance.Amount, receiverBalance.Amount, transfer.Amount)
		if err != nil {
			return shim.Error(err.Error())
		}

		// Update balances //
		balances[transfer.From+" "+transfer.Token] = senderBalance
		balances[transfer.To+" "+transfer.Token] = receiverBalance
		transactions[transfer.Id] = transfer
	}

	// Update States of all the users that did some transaction //
	for _, balance := range balances {
		err = t.updateBalance(stub, balance)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// Prepare output object with updates //
	return generateOutput(balances, nil, transactions)
}

// /* -------------------------------------------------------------------------------------------------
// spendFunds: This function is called when a user wants to spend funds with some tokens. This function
// 			should be called with the following arguments:
// PublicId           string      // ID of the fund spender
// ProviderId         string      // ID of the provider
// Token              string      // Symbol of token to transfer
// Amount             float64     // Amount desired to spend
// ------------------------------------------------------------------------------------------------- */

// func (t *CoinBalanceSmartContract) spendFunds( stub shim.ChaincodeStubInterface,
// 										       args []string ) pb.Response {
// 	//// Retrieve information from the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: GETHISTORY FUNCTION SHOULD BE CALLED " +
// 						   "WITH ONE ARGUMENT." ) }
// 	spending := Spending{}
// 	json.Unmarshal([]byte(args[0]), &spending)

// 	// Retrieve mutiwallet of spender //
// 	walletBytes, err1 := stub.GetState( IndexWallet + spending.PublicId )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: RETRIEVING MULTIWALLET FOR USER: " + spending.PublicId +
// 						   ". " + err1.Error() ) }
// 	spender_wallet := MultiWallet{}
// 	json.Unmarshal(walletBytes, &spender_wallet)
// 	spender_balance := spender_wallet.Balances[spending.Token]

// 	// Retrieve mutiwallet of provider //
// 	walletBytes2, err2 := stub.GetState( IndexWallet + spending.ProviderId )
// 	if err2 != nil {
// 		return shim.Error( "ERROR: RETRIEVING MULTIWALLET FOR USER: " + spending.ProviderId +
// 						   ". " + err2.Error() ) }
// 	provider_wallet := MultiWallet{}
// 	json.Unmarshal(walletBytes2, &provider_wallet)
// 	provider_balance := provider_wallet.Balances[spending.Token]

// 	// Get all PRIVI Loans of user //
// 	PRIVI_credits := []PRIVIloan{}
// 	total_credit := 0.
// 	total_credit_discount := 0.
// 	for priviId, _ := range(spender_balance.PRIVIcreditBorrow) {
// 		privi_loan := PRIVIloan{}
// 		chainCodeArgs := util.ToChaincodeArgs( "getPRIVIcredit", priviId )
// 		response := stub.InvokeChaincode( PRIVI_CREDIT_CHAINCODE, chainCodeArgs,
// 											CHANNEL_NAME )
// 		if response.Status != shim.OK {
// 			return shim.Error( "ERROR INVOKING THE PRIVICREDIT CHAINCODE TO " +
// 								"GET THE PRIVI CREDIT INFO OF: " + priviId ) }
// 		json.Unmarshal(response.Payload, &privi_loan)
// 		PRIVI_credits = append( PRIVI_credits, privi_loan )
// 		borrower_credit := privi_loan.State.Borrowers[ spending.PublicId ]
// 		total_credit = total_credit + borrower_credit.Amount
// 		total_credit_discount = total_credit_discount +
// 		                        borrower_credit.Amount * (1-privi_loan.P_premium) }

// 	transactions := []Transfer{}
// 	PREMIUMS := map[string]Premium{}
// 	// Check if user needs to user PRIVI Credit Funds //
// 	amount_without_credit := math.Max(spender_balance.Amount - total_credit, 0.)
// 	if spending.Amount > amount_without_credit {
// 		if amount_without_credit + total_credit_discount < spending.Amount {
// 			return shim.Error( "ERROR: THE SPENDER " + spending.PublicId + " DOES NOT HAVE " +
// 							   "ENOUGH FUNDS TO SPEND.") }
// 		// Charge first the amount without credit //
// 		transfer := Transfer{
// 			Type: "spending", Token: spending.Token, From: spending.PublicId,
// 			To: spending.ProviderId, Amount: amount_without_credit }
// 		transactions = append( transactions, transfer )
// 		// Credit needed to pay remaining //
// 		credit_needed := spending.Amount - amount_without_credit
// 		for i, priviLoan := range(PRIVI_credits) {
// 			if credit_needed <= 0. {break;}
// 			spender_PRIVI, _ := priviLoan.State.Borrowers[ spending.PublicId ]
// 			// Check if it is enough with this credit //
// 			premium_charged := credit_needed * priviLoan.P_premium
// 			amount_charged := credit_needed + premium_charged
// 			if spender_PRIVI.Amount < amount_charged {
// 				premium_charged = spender_PRIVI.Amount * priviLoan.P_premium
// 				amount_charged = spender_PRIVI.Amount
// 			}
// 			credit_needed = credit_needed - (amount_charged-premium_charged)
// 			transfer_premium := Transfer{
// 				Type: "PRIVI_premium", Token: spending.Token, From: spending.PublicId,
// 				To: "PRIVI Pool " + priviLoan.LoanId, Amount: premium_charged }
// 			transactions = append( transactions, transfer_premium )
// 			// Update PRIVI Loan with new information //
// 			premium_list := priviLoan.State.PremiumList
// 			premium_id := xid.New().String()
// 			premium_list[premium_id] = Premium{ ProviderId: spending.ProviderId,
// 				PremiumId: premium_id, Risk_Pct: 0., Premium_Amount: premium_charged }
// 			//new_premium, _ := json.Marshal(premium_list[premium_id])
// 			PREMIUMS[priviLoan.LoanId] = premium_list[premium_id]
// 			spender_PRIVI.Amount = spender_PRIVI.Amount - amount_charged
// 			spender_balance.Amount = spender_balance.Amount - premium_charged
// 			priviLoan.State.Total_Premium = priviLoan.State.Total_Premium + premium_charged
// 			priviLoan.State.Total_Coverage = priviLoan.State.Total_Coverage + premium_charged
// 			priviLoan.State.PremiumList = premium_list
// 			priviLoan.State.Borrowers[ spending.PublicId ] = spender_PRIVI
// 			PRIVI_credits[i] = priviLoan }
// 	}

// 	// Transfer funds from spender to provider //
// 	spender_balance.Amount = spender_balance.Amount - spending.Amount
// 	provider_balance.Amount = provider_balance.Amount + spending.Amount
// 	transfer := Transfer{
// 		Type: "spending", Token: spending.Token, From: spending.PublicId,
// 		To: spending.ProviderId, Amount: spending.Amount }
// 	transactions = append( transactions, transfer )

// 	// Update state of wallet for both Spender on Blockchain //
// 	spender_wallet.Balances[spending.Token] = spender_balance
// 	spender_wallet.Transaction = transactions
// 	spenderAsBytes, _ := json.Marshal(spender_wallet)
// 	err3 := stub.PutState(IndexWallet + spending.PublicId, spenderAsBytes)
// 	if err3 != nil {
// 		return shim.Error( "ERROR: STORING SPENDER ON BLOCKCHAIN " + err3.Error() ) }

// 	// Update state of wallet for both Provider on Blockchain //
// 	provider_wallet.Balances[spending.Token] = provider_balance
// 	provider_wallet.Transaction = transactions
// 	providerAsBytes, _ := json.Marshal(provider_wallet)
// 	err4 := stub.PutState(IndexWallet + spending.ProviderId, providerAsBytes)
// 	if err4 != nil {
// 		return shim.Error( "ERROR: STORING PROVIDER ON BLOCKCHAIN " + err4.Error() ) }

// 	// Update PRIVI credits of user by invoking PRIVI Coin Chaincode //
// 	update_privi_credits := []string{"updatePRIVIcredits"}
// 	for _, credit := range(PRIVI_credits) {
// 		creditBytes, _ := json.Marshal( credit )
// 		update_privi_credits = append( update_privi_credits, string(creditBytes)) }
// 	multiChainCodeArgs := ToChaincodeArgs( update_privi_credits )
// 	response2 := stub.InvokeChaincode( PRIVI_CREDIT_CHAINCODE, multiChainCodeArgs,
// 										CHANNEL_NAME )
// 	if response2.Status != shim.OK {
// 		return shim.Error( "ERROR INVOKING THE UPDATEMULTIWALLET CHAINCODE TO " +
// 							"UPDATE THE WALLET OF USERs" )
// 	}
// 	premiumsBytes, _ := json.Marshal(PREMIUMS)
// 	return shim.Success(premiumsBytes)
// }

// /* -------------------------------------------------------------------------------------------------
// getHistory: This function returns the transaction history of a given Public Key Id from a given
//             timestamp. Args: is an array containing a json with the following attributes:
// PublicId           string   // Id of the user to retrieve the history
// Timestamp          string   // Timestamp of time from which retrieve the history
// ------------------------------------------------------------------------------------------------- */

// func (t *CoinBalanceSmartContract) getHistory( stub shim.ChaincodeStubInterface,
// 	                                           args []string ) pb.Response {

// 	// Retrieve information from the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: GETHISTORY FUNCTION SHOULD BE CALLED " +
// 						   "WITH ONE ARGUMENT." )
// 	}
// 	history := History{}
// 	input := []byte(args[0])
// 	json.Unmarshal(input, &history)

// 	// Retrieve iterator of History for the User Public Key //
// 	resultsIterator, err1 := stub.GetHistoryForKey(IndexWallet + history.PublicId)
// 	if err1 != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE HISTORY FROM BLOCKCHAIN. " +
// 						   "ERROR WAS: " + err1.Error())
// 	}
// 	defer resultsIterator.Close()

// 	// Create response. Filter for timestamps greater than the given one //
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")
// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		response, err2 := resultsIterator.Next()
// 		if err2 != nil {
// 			return shim.Error( "ERROR: GETTING NEXT ITERATOR. " +
// 						       "ERROR WAS: " + err2.Error())
// 		}
// 		//Check if txn time is greater than the query time //
// 		time_txn :=time.Unix( response.Timestamp.Seconds,
// 			                  int64(response.Timestamp.Nanos))
// 		if ( time_txn.Unix() < history.Timestamp ||
// 		     response.IsDelete) {continue;}

// 		// Add a comma before array members, suppress it for first one
// 		if bArrayMemberAlreadyWritten == true {buffer.WriteString(",")}
// 		buffer.WriteString("{\"TxId\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(response.TxId)
// 		buffer.WriteString("\"")

// 		data := make(map[string]interface{})
// 		json.Unmarshal(response.Value, &data)
// 		buffer.WriteString(", \"Value\":")
// 		dataAsBytes, _ := json.Marshal(data["Transaction"])
// 		buffer.WriteString(string(dataAsBytes))

// 		buffer.WriteString(", \"Timestamp\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString( strconv.FormatInt(time_txn.Unix(),10) )
// 		buffer.WriteString("\"")
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	return shim.Success(buffer.Bytes())
// }

/* -------------------------------------------------------------------------------------------------
mint: This function is called to perform a swapping of user's token wallets by the same amount of
	     Fabric tokens version. It mints new tokens and transfer it to the user balance.
		 Args: is an array containing a json with the following attributes
Type               string   // Type of mint
Token              string   // Symbol of the token to swap
From               string   // From minting (Ethereum, Pod..)
To                 string   // Id of the receiver of the tokens
Amount             float64  // Amount of tokens to mint
Id                 string   // ID of the transaction
Date               float64  // Date timestamp
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) mint(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {
	// Retrieve information from the input //
	if len(args) != 1 {
		return shim.Error("ERROR: MINT FUNCTION SHOULD BE CALLED " +
			"WITH ONE ARGUMENT.")
	}
	input := Transfer{}
	json.Unmarshal([]byte(args[0]), &input)
	transactions := make(map[string]Transfer)
	balances := make(map[string]Balance)

	// Get state of the token from the Ledger //
	token, err := t.getToken(stub, input.Token)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Retrieve user balance //
	userBalance, err2 := t.checkBalance(stub, input.To, input.Token, true)
	if err2 != nil {
		return shim.Error(err2.Error())
	}

	// Transfer swapping amount to user //
	userBalance.Amount += input.Amount
	balances[input.To+" "+input.Token] = userBalance
	transactions[input.Id] = input

	// Update user balance //
	err = t.updateBalance(stub, userBalance)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Mint amount of tokens in the system and update state //
	token.Supply += input.Amount
	err = t.updateToken(stub, token)
	if err != nil {
		return shim.Error(err.Error())
	}
	updateTokens := make(map[string]Token)
	updateTokens[token.Symbol] = token

	// Prepare output object with updates //
	return generateOutput(balances, updateTokens, transactions)

}

// /* -------------------------------------------------------------------------------------------------
// multiMint: This function is called to perform a minting of a particular token to a
//            given list of addresses.
// Token           string                     // Token to be minted
// FromAddress     string                     // Address from with the minting is generated
// Type            string                     // Type of transaction
// TxnId           string                     // Id of the transactions
// Date            string                     // Date of the transfer
// TotalAmount     float64                    // Total Amount to mint
// Transfers       map[string]float64         // List of transfer to do form minting
// ------------------------------------------------------------------------------------------------- */

// func (t *CoinBalanceSmartContract) multiMint(stub shim.ChaincodeStubInterface,
// 	args []string) pb.Response {

// 	// Retrieve information from the input //
// 	if len(args) != 1 {
// 		return shim.Error("ERROR: MINT FUNCTION SHOULD BE CALLED " +
// 			"WITH 1 ARGUMENTS.")
// 	}
// 	input := MultiMinter{}
// 	json.Unmarshal([]byte(args[0]), &input)

// 	// Get state of the token from the Ledger //
// 	token, err := t.getToken(stub, input.Token)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	// Check if transfer is allowed //
// 	err = t.checkTokenTransferConditions(stub, input.Token,
// 		input.Date, input.TotalAmount)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	transactions := make(map[string]Transfer)
// 	balances := make(map[string]Balance)
// 	transfer := {
// 		Type: input.Type, From: input.AddressFrom, Date: input.Date
// 	}

// 	// Iterate throught all the transactions on the multitransfer call //
// 	totalAmount := 0.
// 	txnNum := 0
// 	for addressTo, amount := range input.Transfers {
// 		totalAmount += amount
// 		txnId := input.TxnId + strconv.Itoa(txnNum)
// 		// Check if 0 amount or equal address //
// 		if input.FromAddress == addressTo || amount == 0 {
// 			continue
// 		}
// 		// Check if receiver is already in transaction users list //
// 		receiverBalance, inList = balances[addressTo+" "+input.Token]
// 		if !inList {
// 			receiverBalance, err = t.checkBalance(stub, addressTo, input.Token, true)
// 			if err != nil {
// 				return shim.Error(err.Error())
// 			}
// 		}
// 		// Check that the sender holds the amount to send and transfer funds //
// 		receiverBalance.Amount += amount
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 		// Generate Transfer object //
// 		transfer.To = addressTo
// 		transfer.Amount = amount
// 		transfer.Id = txnId

// 		// Update balance of Receiver //
// 		balances[addressTo+" "+input.Token] = receiverBalance
// 		transactions[txnId] = transfer
// 		++txnNum
// 	}

// 	if math.Abs(input.TotalAmount-totalAmount) < PRECISSION {
// 		return shim.Error( "ERROR: TOTAL AMOUNT MINTED TO USERS SHOULD BE EQUAL " +
// 			"TO THE TOTAL AMOUNT DESIRED TO MINT." )
// 	}

// 	// Update States of all the users that did some transaction //
// 	for _, balance := range balances {
// 		err = t.updateBalance(stub, balance)
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}
// 	}

// 	// Mint amount of tokens in the system and update state //
// 	token.Supply += totalAmount
// 	err = t.updateToken(stub, token)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	tokens := make( map[string]Token )
// 	tokens[token.Symbol] = token

// 	// Prepare output object with updates //
// 	outputBytes, errOut := t.generateOutput(stub, balances, tokens, transactions)
// 	if errOut != nil {
// 		return shim.Error(errOut.Error())
// 	}
// 	return shim.Success(outputBytes)
// }

/* -------------------------------------------------------------------------------------------------
burn: This function is called to perform a swapping back of user's token in Fabric version by the
          same amount on original token version. It burns the Fabric tokens of the user.
	  	  Args: is an array containing a json with the following attributes
Type               string   // Type of mint
Token              string   // Symbol of the token to swap
From               string   // From minting (Ethereum, Pod..)
To                 string   // Id of the receiver of the tokens
Amount             float64  // Amount of tokens to mint
TxnId              string   // ID of the transaction
Date               float64  // Date timestamp
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) burn(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {
	// Retrieve information from the input //
	if len(args) != 1 {
		return shim.Error("ERROR: WITHDRAW FUNCTION SHOULD BE CALLED " +
			"WITH ONE ARGUMENT.")
	}
	input := Transfer{}
	json.Unmarshal([]byte(args[0]), &input)
	transactions := make(map[string]Transfer)
	balances := make(map[string]Balance)

	// Get state of the token from the Ledger //
	token, err := t.getToken(stub, input.Token)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Retrieve user balance //
	userBalance, err2 := t.checkBalance(stub, input.From, input.Token, true)
	if err2 != nil {
		return shim.Error(err2.Error())
	}

	// Burn amount from user //
	userBalance.Amount, err = saveSubstraction(userBalance.Amount, input.Amount)
	if err != nil {
		return shim.Error(err.Error())
	}
	balances[input.From+" "+input.Token] = userBalance
	transactions[input.Id] = input

	// Update user balance //
	err = t.updateBalance(stub, userBalance)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Burn amount of tokens in the system and update state //
	token.Supply, err = saveSubstraction(token.Supply, input.Amount)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = t.updateToken(stub, token)
	if err != nil {
		return shim.Error(err.Error())
	}
	updateTokens := make(map[string]Token)
	updateTokens[token.Symbol] = token

	// Prepare output object with updates //
	return generateOutput(balances, updateTokens, transactions)
}

/* -------------------------------------------------------------------------------------------------
getBalancesOfAddress: this function retrieves the balances of an address
address               string    // address to get balances
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getBalancesOfAddress(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	balances, err := findAllBalacesOfAddress(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	outputBytes, _ := json.Marshal(balances)
	return shim.Success(outputBytes)
}

/* -------------------------------------------------------------------------------------------------
getBalancesOfTokenHolders: this function retrieves the balances of an address
address               string    // address to get balances
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getBalancesOfTokenHolders(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	balances, err := findAllHoldersOfToken(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	outputBytes, _ := json.Marshal(balances)
	return shim.Success(outputBytes)
}

/* -------------------------------------------------------------------------------------------------
updateTokenInfo: this function updates the information of a given Token already registered in the
                 system (keeping the supply)/
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) updateTokenInfo(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	token := Token{}
	err := json.Unmarshal([]byte(args[0]), &token)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Get state of the token from the Ledger //
	var tokenOld Token
	tokenOld, err = t.getToken(stub, token.Symbol)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Keep Supply //
	token.Supply = tokenOld.Supply
	err = t.updateToken(stub, token)
	if err != nil {
		return shim.Error(err.Error())
	}
	updateTokens := make(map[string]Token)
	updateTokens[token.Symbol] = token

	return generateOutput(nil, updateTokens, nil)
}

/* -------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------- */

func main() {
	err := shim.Start(&CoinBalanceSmartContract{})
	if err != nil {
		fmt.Errorf("Error starting Token chaincode: %s", err)
	}
}
