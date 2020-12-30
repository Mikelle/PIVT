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

	//"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

/* -------------------------------------------------------------------------------------------------
getActor:  this function returns the supply and information of a given actor
------------------------------------------------------------------------------------------------- */

func getActor(stub shim.ChaincodeStubInterface, publicId string) (Actor, error) {

	// Initialise new empty balance for the token //
	actor := Actor{PublicId: publicId}

	// Check if Balance is already registered on Blockchain //
	isLoaded, err := actor.LoadState(stub)
	if err != nil {
		return actor, errors.New("ERROR: CHECKING IF ACTOR IS ALREADY " +
			"REGISTERED. " + err.Error())
	}
	if !isLoaded {
		return actor, errors.New("ERROR: ACTOR " + publicId + " IS NOT REGISTERED " +
			"ON THE SYSTEM. ")
	}
	return actor, nil
}

/* -------------------------------------------------------------------------------------------------
checkActorRegistered:  this function returns the supply and information of a given token
------------------------------------------------------------------------------------------------- */

func checkActorRegistered(stub shim.ChaincodeStubInterface, publicId string) (bool, error) {

	// Initialise new empty balance for the token //
	actor := Actor{PublicId: publicId}

	// Check if Balance is already registered on Blockchain //
	isLoaded, err := actor.LoadState(stub)
	if err != nil {
		return false, errors.New("ERROR: CHECKING IF ACTOR IS ALREADY " +
			"REGISTERED. " + err.Error())
	}
	if !isLoaded {
		return false, nil
	}
	return true, nil
}

/* -------------------------------------------------------------------------------------------------
updateActor:  this function updates/register a actor on blockchain
------------------------------------------------------------------------------------------------- */

func updateActor(stub shim.ChaincodeStubInterface, actor Actor) error {

	// Update actor on Blockchain //
	if err := actor.SaveState(stub); err != nil {
		return err
	}
	return nil
}

/* -------------------------------------------------------------------------------------------------
getRoleList: returns the list of actors with its system info of a given type
------------------------------------------------------------------------------------------------- */

func getRoleList(stub shim.ChaincodeStubInterface, role string) ([]string, error) {
	queryString := fmt.Sprintf(`{"selector":{"Role":"%s"}}`, role)
	it, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New("ERROR: unable to get an iterator over the balances")
	}
	defer it.Close()
	var actorList []string
	for it.HasNext() {
		response, error := it.Next()
		if error != nil {
			message := fmt.Sprintf("unable to get the next element: %s", error.Error())
			return nil, errors.New(message)
		}
		var actor Actor
		if err = json.Unmarshal(response.Value, &actor); err != nil {
			message := fmt.Sprintf("ERROR: unable to parse the response: %s", err.Error())
			return nil, errors.New(message)
		}
		actorList = append(actorList, actor.PublicId)
	}
	return actorList, nil
}

/* -------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------- */
