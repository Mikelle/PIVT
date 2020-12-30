/*--------------------------------------------------------------------------
----------------------------------------------------------------------------
   HELPER FUNCTIONS CALLED SEVERAL TIMES ON MAIN SMART CONTRACT FUNCIOTNS
----------------------------------------------------------------------------
-------------------------------------------------------------------------- */

package main

import (
	//"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/* -------------------------------------------------------------------------------------------------
 checkPermissions: check if user has permissions to call a given function
------------------------------------------------------------------------------------------------- */

func checkPermissions(stub shim.ChaincodeStubInterface, userRole string,
	functionName string) error {

	if err := cid.AssertAttributeValue(stub, "userRole", userRole); err != nil {
		return errors.New("PERMISSION DENIED TO CALL " + functionName)
	}
	return nil

}

/* -------------------------------------------------------------------------------------------------
 validateSignature: this function validates when a user signs it.
------------------------------------------------------------------------------------------------- */

func validateSignature(publicKey string, hash string, signature string) error {

	publicKeyBytes, err := hexutil.Decode(publicKey)
	if err != nil {
		return errors.New("ERROR: ERROR DECODING PUBLIC KEY " + publicKey)
	}
	var hashBytes []byte
	hashBytes, err = hexutil.Decode(hash)
	if err != nil {
		return errors.New("ERROR: ERROR DECODING HASH")
	}
	var signatureBytes []byte
	signatureBytes, err = hexutil.Decode(signature)
	if err != nil {
		return errors.New("ERROR: ERROR DECODING SIGNATURE")
	}

	isVerified := crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes[:len(signatureBytes)-1])
	if !isVerified {
		return errors.New("ERROR: THE SIGNATURE IS NOT VALID. NO PERMISSIONS FOR ADDRESS " +
			publicKey)
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
checkTokenListed:  this function checks if a token is listed.
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) checkTokenListed(stub shim.ChaincodeStubInterface,
	tokenSymbol string) error {

	// Check if token is listed on blockchain //
	tokenBytes, err := stub.GetState(IndexToken + tokenSymbol)
	if err != nil || tokenBytes == nil {
		return errors.New("ERROR: TOKEN " + tokenSymbol + " IS NOT REGISTERED " +
			"ON THE SYSTEM. ")
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
getToken:  this function returns the supply and information of a given token
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getToken(stub shim.ChaincodeStubInterface,
	tokenSymbol string) (Token, error) {

	// Initialise new empty balance for the token //
	token := Token{Symbol: tokenSymbol}

	// Check if Balance is already registered on Blockchain //
	isLoaded, err := token.LoadState(stub)
	if err != nil {
		return token, errors.New("ERROR: CHECKING IF TOKEN IS ALREADY " +
			"REGISTERED. " + err.Error())
	}
	if !isLoaded {
		return token, errors.New("ERROR: TOKEN " + tokenSymbol + " IS NOT REGISTERED " +
			"ON THE SYSTEM. ")
	}
	return token, nil
}

/* -------------------------------------------------------------------------------------------------
checkTokenRegistered:  this function returns the supply and information of a given token
------------------------------------------------------------------------------------------------- */

func checkTokenRegistered(stub shim.ChaincodeStubInterface,
	tokenSymbol string) (bool, error) {

	// Initialise new empty balance for the token //
	token := Token{Symbol: tokenSymbol}

	// Check if Balance is already registered on Blockchain //
	isLoaded, err := token.LoadState(stub)
	if err != nil {
		return false, errors.New("ERROR: CHECKING IF TOKEN IS ALREADY " +
			"REGISTERED. " + err.Error())
	}
	if !isLoaded {
		return false, nil
	}
	return true, nil
}

/* -------------------------------------------------------------------------------------------------
updateToken:  this function updates/register a token on blockchain
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) updateToken(stub shim.ChaincodeStubInterface,
	token Token) error {

	// Update token on Blockchain/ /
	if err := token.SaveState(stub); err != nil {
		return err
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
getTokenInfoByType: returns the list of tokens with its system info of a given type
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getTokenInfoByType(stub shim.ChaincodeStubInterface,
	tokenType string) ([]Token, error) {
	queryString := fmt.Sprintf(`{"selector":{"TokenType":"%s"}}`, tokenType)
	it, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var tokenList []Token
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var token Token
		if err = json.Unmarshal(response.Value, &token); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		tokenList = append(tokenList, token)
	}
	return tokenList, nil
}

/* -------------------------------------------------------------------------------------------------
getTokenListByType: returns the list of tokens with its system info of a given type
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) getTokenListByType(stub shim.ChaincodeStubInterface,
	tokenType string) ([]string, error) {
	queryString := fmt.Sprintf(`{"selector":{"TokenType":"%s"}}`, tokenType)
	it, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var tokenList []string
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var token Token
		if err = json.Unmarshal(response.Value, &token); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		tokenList = append(tokenList, token.Symbol)
	}
	return tokenList, nil
}

/* -------------------------------------------------------------------------------------------------
checkBalance: this function checks if the balance of the user fo a given token is
           created. Otherwise, it create a balance for him.
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) checkBalance(stub shim.ChaincodeStubInterface,
	user string, token string, check bool) (Balance, error) {

	// Initialise new empty balance for the token //
	balance := Balance{
		Address: user, Token: token}

	// Check if address is registered //
	if check {
		walletExist := t.checkAddressExist(stub, user)
		if !walletExist {
			return balance, errors.New("ERROR: THE ADDRESS FOR " + user +
				" IS NOT REGISTERED.")
		}
	}

	// Check if Balance is already registered on Blockchain //
	isLoaded, err := balance.LoadState(stub)
	if err != nil {
		return balance, errors.New("ERROR: CHECKING IF BALANCE IS ALREADY " +
			"REGISTERED. " + err.Error())
	}
	if !isLoaded {
		balance = Balance{
			Address: user, Token: token,
			Amount: 0., Credit: 0.}
	}

	return balance, nil
}

/* -------------------------------------------------------------------------------------------------
updateBalance: this function updates the balance of a user.
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) updateBalance(stub shim.ChaincodeStubInterface,
	balance Balance) error {
	if err := balance.SaveState(stub); err != nil {
		return err
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
checkUserExist: this function checks taht an user is registered
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) checkUserExist(stub shim.ChaincodeStubInterface,
	user string) error {
	// Register financial scores for the user //
	invoke_call := []string{"getUser"}
	invoke_call = append(invoke_call, user)
	multiChainCodeArgs := ToChaincodeArgs(invoke_call)
	response := stub.InvokeChaincode(DATA_PROTOCOL_CHAINCODE, multiChainCodeArgs,
		CHANNEL_NAME)
	if response.Status != shim.OK {
		return errors.New("ERROR: " + user + " IS NOT REGISTED ON BLOCKCHAIN.")
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
checkAddressExist: this function checks taht an user is registered
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) checkAddressExist(stub shim.ChaincodeStubInterface,
	wallet string) bool {

	result, err := stub.GetState(IndexWallets + wallet)
	if err != nil || result == nil {
		return false
	}
	return true
}

/* -------------------------------------------------------------------------------------------------
generateOutput: this function generates the output.
------------------------------------------------------------------------------------------------- */

func generateOutput(balances map[string]Balance, tokens map[string]Token,
	transactions map[string]Transfer) pb.Response {

	output := Output{UpdateBalances: balances,
		UpdateTokens: tokens,
		Transactions: transactions}
	outputBytes, err := json.Marshal(output)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(outputBytes)
}

/* -------------------------------------------------------------------------------------------------
transferHelper: this function computes a transfer given balances and returns updated balances
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) transferHelper(stub shim.ChaincodeStubInterface,
	senderBalance float64, receiverBalance float64, amount float64) (float64, float64, error) {

	err := errors.New("")
	// Substract fund from sender
	senderBalance, err = saveSubstraction(senderBalance, amount)
	if err != nil {
		return senderBalance, receiverBalance, err
	}

	// Add funds to receiver
	receiverBalance, err = saveAddition(receiverBalance, amount)
	if err != nil {
		return senderBalance, receiverBalance, err
	}

	return senderBalance, receiverBalance, nil

}

/* -------------------------------------------------------------------------------------------------
 getTokenHolderList: returns the balances of the holders of a token
------------------------------------------------------------------------------------------------- */

func getTokenHolderList(stub shim.ChaincodeStubInterface, token string) ([]string, error) {
	queryString := fmt.Sprintf(`{"selector":{"Token":"%s"}}`, token)
	it, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var holderList []string
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var balance Balance
		if err = json.Unmarshal(response.Value, &balance); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		holderList = append(holderList, balance.Address)
	}
	return holderList, nil
}

/* -------------------------------------------------------------------------------------------------
findAllHoldersOfToken: returns the balances of the holders of a token
------------------------------------------------------------------------------------------------- */

func findAllHoldersOfToken(stub shim.ChaincodeStubInterface, token string) ([]Balance, error) {
	queryString := fmt.Sprintf(`{"selector":{"Token":"%s"}}`, token)
	it, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var balances []Balance
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var balance Balance
		if err = json.Unmarshal(response.Value, &balance); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		balances = append(balances, balance)
	}
	return balances, nil
}

/* -------------------------------------------------------------------------------------------------
findAllBalacesOfAddress: gives the balances of a user
------------------------------------------------------------------------------------------------- */

func findAllBalacesOfAddress(stub shim.ChaincodeStubInterface, address string) ([]Balance, error) {
	it, err := stub.GetStateByPartialCompositeKey(IndexBalances, []string{address})
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var balances []Balance
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var balance Balance
		if err = json.Unmarshal(response.Value, &balance); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		balances = append(balances, balance)
	}
	return balances, nil
}

/* -------------------------------------------------------------------------------------------------
checkTokenTransferConditions: this function check if transfer with a given token are allowed
------------------------------------------------------------------------------------------------- */

func (t *CoinBalanceSmartContract) checkTokenTransferConditions(stub shim.ChaincodeStubInterface, tokenSymbol string,
	date int64, amount float64) error {

	// Retrieve token //
	token, err := t.getToken(stub, tokenSymbol)
	if err != nil {
		return err
	}

	// Get token conditions //
	if date < token.LockUpDate {
		return errors.New("ERROR: THE TOKEN CANNOT BE TRANSFERED YET. IT IS " +
			" IN LOCK-UP PERIOD.")
	}

	// Check if integer in case of NFT token //
	if token.TokenType == NFT_POD_TOKEN {
		if amount != math.Trunc(amount) {
			return errors.New("ERROR: THE TRANSFER AMOUNT FOR A NFT POD TOKEN " +
				" SHOULD BE AN INTEGER.")
		}
	}
	return nil
}
