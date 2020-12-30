package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	//"github.com/rs/xid"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	//"github.com/hyperledger/fabric/common/util"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/* -------------------------------------------------------------------------------------------------
Init:  this function register Cache as the Admin of the Network at the deployment of the
       Cache Blockchain. Args: array containing a string:
PrivateKeyID           string   // Private Key of the admin of the smart contract
------------------------------------------------------------------------------------------------- */

func (t *DataProtocolSmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {

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

func (t *DataProtocolSmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// Retrieve function and arguments //
	function, args := stub.GetFunctionAndParameters()

	if err := cid.AssertAttributeValue(stub, "userRole", ADMIN_ROLE); err != nil {
		message := "PERMISSION DENIED TO CALL DATA PROTOCOL"
		return pb.Response{Status: 403, Message: message}
	}

	// Call the proper function //
	switch function {
	case "register":
		return t.register(stub, args)
	case "attachAddress":
		return t.attachAddress(stub, args)
	case "getUser":
		actor, err := getActor(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		actorBytes, _ := json.Marshal(actor)
		return shim.Success(actorBytes)
	case "getRoleList":
		actorList, err := getRoleList(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		actorListBytes, _ := json.Marshal(actorList)
		return shim.Success(actorListBytes)

		// case "getRoleList":
		// 	return t.getRoleList(stub, args)
		// case "getPrivacy":
		// 	return t.getPrivacy(stub,args)
		// case "modifyPrivacy":
		// 	return t.modifyPrivacy(stub,args)
		// case "getBusinessList":
		// 	return t.getBusinessList(stub,args)

		// case "getUserList":
		// 	return t.getUserList(stub,args)
		// case "getTIDList":
		// 	return t.getTIDList(stub,args)
		// case "encryptData":
		// 	return t.encryptData(stub, args)
		// case "decryptData":
		// 	return t.decryptData(stub, args)
		// case "insightDiscovery":
		// 	return t.insightDiscovery(stub, args)
		// case "insightPurchase":
		// 	return t.insightPurchase(stub, args)
		// case "insightTarget":
		// 	return t.insightTarget(stub, args)
	}
	return shim.Error("Incorrect function name: " + function)
}

/* -------------------------------------------------------------------------------------------------
register:  temporarily register funciton
PublicId            string    // Public identifier of the user
Role                string    // Role of the actor in the Cache Ecosystem
------------------------------------------------------------------------------------------------- */

func (t *DataProtocolSmartContract) register(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Retrieving the actor information for registration //
	actor := Actor{}
	err := json.Unmarshal([]byte(args[0]), &actor)
	if err != nil {
		return shim.Error("ERROR: GETTING INPUT INFORMATION. " +
			err.Error())
	}

	// Get role of user //
	scores := FinancialScores{}
	switch actor.Role {
	case USER_ROLE:
		scores.EndorsementScore = 0.5
		scores.TrustScore = 0.5
	case BUSINESS_ROLE:
		scores.EndorsementScore = 0.7
		scores.TrustScore = 0.7
	case GUARANTOR_ROLE:
		scores.EndorsementScore = 0.85
		scores.TrustScore = 0.85
	case COURTMEMBER_ROLE:
		scores.EndorsementScore = 0.9
		scores.TrustScore = 0.9
	case EXCHANGE_ROLE:
		scores.EndorsementScore = 0.7
		scores.TrustScore = 0.7
	case ADMIN_ROLE:
		scores.EndorsementScore = 1.
		scores.TrustScore = 1.
	}

	// Register actor on blockchain //
	err = updateActor(stub, actor)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Register financial scores for the user //
	invoke_call := []string{"initialiseFinancialScores"}
	invoke_call = append(invoke_call, actor.PublicId, toStringMethod(scores))
	multiChainCodeArgs := ToChaincodeArgs(invoke_call)
	response := stub.InvokeChaincode(COIN_BALANCE_CHAINCODE, multiChainCodeArgs,
		CHANNEL_NAME)
	if response.Status != shim.OK {
		return shim.Error("ERROR CREATING SCORES FOR " + actor.PublicId +
			" ON BLOCKCHAIN. " + response.Message)
	}

	return shim.Success(nil)
}

/* -------------------------------------------------------------------------------------------------
attachAddress: attach new wallet to user
PublicId               string    // Public identifier of the user (args[0])
PublicAddress          string    // Public address of the user (args[1])
------------------------------------------------------------------------------------------------- */

func (t *DataProtocolSmartContract) attachAddress(stub shim.ChaincodeStubInterface,
	args []string) pb.Response {

	// Load the user from the UserId //
	actor, err := getActor(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	// Register address for the user //
	invoke_call := []string{"registerAddress"}
	invoke_call = append(invoke_call, args[1])
	multiChainCodeArgs := ToChaincodeArgs(invoke_call)
	response := stub.InvokeChaincode(COIN_BALANCE_CHAINCODE, multiChainCodeArgs,
		CHANNEL_NAME)
	if response.Status != shim.OK {
		return shim.Error("ERROR ATTACHING ADDRESS FOR " + args[0] +
			" ON BLOCKCHAIN. " + response.Message)
	}

	// Attach Public Address //
	actor.PublicAddress = args[1]
	err = updateActor(stub, actor)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

/* --------------------------------------------------------------------------------------------------
getRoleList:  this function returns the list of users registered in the PRIVI Blockchain. This
              function does not need to be called with any argument.
------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) getRoleList(stub shim.ChaincodeStubInterface,
// 	args []string) pb.Response {

// 	// Retrieve list of users from the Blockchain //
// 	var err error
// 	roleList := []byte{}
// 	switch args[0] {
// 	case USER_ROLE:
// 		roleList, err = stub.GetState(IndexUserList)
// 	case BUSINESS_ROLE:
// 		roleList, err = stub.GetState(IndexBusinessList)
// 	case GUARANTOR_ROLE:
// 		roleList, err = stub.GetState(IndexGuarantorList)
// 	case COURTMEMBER_ROLE:
// 		roleList, err = stub.GetState(IndexCourtMemberList)
// 	case EXCHANGE_ROLE:
// 		roleList, err = stub.GetState(IndexExchangesList)
// 	}
// 	if err != nil {
// 		return shim.Error("ERROR: RETRIEVING THE ROLE LIST")
// 	}

// 	return shim.Success(roleList)
// }

// /* -------------------------------------------------------------------------------------------------
// register:  this function registers a new actor in the Blockchain. It first checks that the actor is
// 		   not registered and then calls the Cache Coin and Base Coin smart contracts to initialise
// 		   its balance to 0. Args: array contaning a json with:
// PublicId            string    // Public identifier of the suer
// Role                string    // Role of the actor in the Cache Ecosystem
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) register( stub shim.ChaincodeStubInterface,
// 	  										  args []string) pb.Response {

// 	// Retrieving the actor information for registration //
// 	actor := Actor{}
// 	err1 := json.Unmarshal( []byte(args[0]), &actor )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: GETTING INPUT INFORMATION. " +
// 	                       err1.Error() ) }
// 	actor.TargetTIDs = make( map[string]bool )

// 	// Check that the actor is not already registered in the Network //
// 	callerBytes, err2 := stub.GetState(IndexNetwork+actor.PublicId)
// 	if err2 != nil {
// 		return shim.Error( "ERROR: FAILED TO GET CALLER. " +
// 	                        err2.Error() )
// 	} else if callerBytes != nil {
// 		return shim.Error( "ERROR: ACTOR " + actor.PublicId +
// 		                   " ALREADY REGISTERED IN CACHE NETWORK." ) }

// 	// Generate DID that will be used for encrypt and decrypt data //
// 	encryption_DID := xid.New().String() + xid.New().String() +
// 					  xid.New().String() + xid.New().String()
// 	err3 := stub.PutState( IndexEncryption+actor.PublicId,
// 						   []byte(encryption_DID) )
// 	err4 := stub.PutState( IndexDecryption+encryption_DID,
// 						   []byte(actor.PublicId) )
// 	if (err3 != nil || err4 != nil) {
// 		return shim.Error( "ERROR ENCRYPTING ACTOR " + actor.PublicId +
// 		                   " ON BLOCKCHAIN." ) }

// 	// Register a Wallet for the new actor //
// 	invoke_call := []string{ "updateScores" }
// 	invoke_call = append(invoke_call, actor.PublicId)
// 	multiChainCodeArgs := ToChaincodeArgs( invoke_call )
// 	response := stub.InvokeChaincode( COIN_BALANCE_CHAINCODE, multiChainCodeArgs,
// 		                              CHANNEL_NAME )
// 	if response.Status != shim.OK {
// 		return shim.Error( "ERROR CREATING BALANCER FOR " + actor.PublicId +
// 						   " ON BLOCKCHAIN. " + response.Message ) }
// 	output := Output{}
// 	json.Unmarshal( response.Payload, &output )

// 	// Retrieve list of all businesses registered in Blockchain //
// 	businessListBytes, err5 := stub.GetState( IndexBusinessList )
// 	if err5 != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE BUSINESS LIST. " +
// 		                    err5.Error() ) }
// 	business_list := []string{}
// 	json.Unmarshal( businessListBytes, &business_list )

// 	// Retrieve list of all users registered in Blockchain //
// 	userListBytes, err6 := stub.GetState( IndexUserList )
// 	if err6 != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE USER LIST. " +
// 		                    err6.Error() ) }
// 	user_list := []string{}
// 	json.Unmarshal(userListBytes, &user_list)

// 	// If the actor is an user, initialise all the privacies to true //
// 	if actor.Role == USER_ROLE {
// 		privacy := make( map[string]bool )
// 		for i, _ := range business_list {
// 			id := business_list[i]
// 			privacy[id] = true }
// 		actor.Privacy = privacy
// 		// Add new user to the user list of the Smart Contract //
// 		user_list = append(user_list, actor.PublicId)
// 		userListBytes2, _ := json.Marshal(user_list)
// 		err7 := stub.PutState( IndexUserList, userListBytes2 )
// 		if err7 != nil {
// 			return shim.Error( "ERROR: ADDING USER " + actor.PublicId +
// 							   " TO THE USER LIST" + err7.Error() )
// 		}
// 	// If the actor is a business, initialise privacy for all users to true //
// 	} else if actor.Role == BUSINESS_ROLE {
// 		for i, _ := range user_list {
// 			user := Actor{}
// 			userBytes, err7 := stub.GetState( IndexNetwork+user_list[i] )
// 			json.Unmarshal(userBytes, &user)
// 			if err7 != nil {
// 				return shim.Error( "ERROR: GETTING USER INFORMATION FOR " +
// 									user_list[i] + ". " + err1.Error() ) }
// 			privacy := user.Privacy
// 			privacy[actor.PublicId] = true
// 			user.Privacy = privacy
// 			userBytes2, _ := json.Marshal(user)
// 			err8 := stub.PutState( IndexNetwork+user_list[i], userBytes2 )
// 			if err8 != nil {
// 				return shim.Error( "ERROR UPDATING USER " + actor.PublicId +
// 								   " ON BLOCKCHAIN." + err8.Error() )
// 			}
// 		}
// 		// Add new business to business list of Smart Contract //
// 		business_list = append(business_list, actor.PublicId)
// 		businessListBytes2, _ := json.Marshal(business_list)
// 		err9 := stub.PutState( IndexBusinessList, businessListBytes2 )
// 		if err9 != nil {
// 			return shim.Error( "ERROR: ADDING BUSINESS " + actor.PublicId +
// 							   " TO THE BUSINESS LIST" )
// 		}
// 	// If the actor is a Guarantor, add it to list //
// 	} else if actor.Role == GUARANTOR_ROLE {
// 		// Retrieve list of all guarantors registered in Blockchain //
// 		guarantorListBytes, err10 := stub.GetState( IndexGuarantorList )
// 		if err10 != nil {
// 			return shim.Error( "ERROR: RETRIEVING THE GUARANTOR LIST. " +
// 								err10.Error() ) }
// 		guarantor_list := []string{}
// 		json.Unmarshal( guarantorListBytes, &guarantor_list )
// 		// Add new guarantor to guarantors list of Smart Contract //
// 		guarantor_list = append(guarantor_list, actor.PublicId)
// 		guarantorListBytes2, _ := json.Marshal(guarantor_list)
// 		err11 := stub.PutState( IndexGuarantorList, guarantorListBytes2 )
// 		if err11 != nil {
// 			return shim.Error( "ERROR: ADDING GUARANTOR " + actor.PublicId +
// 							   " TO THE BUARANTOR LIST" ) }
// 	// If the actor is a Digital Court Member, add it to list //
// 	} else if actor.Role == COURTMEMBER_ROLE {
// 		// Retrieve list of all court members registered in Blockchain //
// 		courtmemberListBytes, err12 := stub.GetState( IndexCourtMemberList )
// 		if err12 != nil {
// 			return shim.Error( "ERROR: RETRIEVING THE COURT MEMBER LIST. " +
// 								err12.Error() ) }
// 		court_member_list := []string{}
// 		json.Unmarshal( courtmemberListBytes, &court_member_list )
// 		// Add new court member to court member list of Smart Contract //
// 		court_member_list = append(court_member_list, actor.PublicId)
// 		courtmemberListBytes2, _ := json.Marshal(court_member_list)
// 		err13 := stub.PutState( IndexCourtMemberList, courtmemberListBytes2 )
// 		if err13 != nil {
// 			return shim.Error( "ERROR: ADDING COURT MEMEBER " + actor.PublicId +
// 							   " TO THE COURT MEMBER LIST" ) }
// 	// If the actor is an Exchange, add it to list //
// 	} else if actor.Role == EXCHANGE_ROLE {
// 		// Retrieve list of all exchanges registered in Blockchain //
// 		exchangeListBytes, err14 := stub.GetState( IndexExchangesList )
// 		if err14 != nil {
// 			return shim.Error( "ERROR: RETRIEVING THE EXCHANGE LIST. " +
// 								err14.Error() ) }
// 		exchange_list := []string{}
// 		json.Unmarshal( exchangeListBytes, &exchange_list )
// 		// Add new business to business list of Smart Contract //
// 		exchange_list = append(exchange_list, actor.PublicId)
// 		exchangeListBytes2, _ := json.Marshal(exchange_list)
// 		err15 := stub.PutState( IndexExchangesList, exchangeListBytes2 )
// 		if err15 != nil {
// 			return shim.Error( "ERROR: ADDING EXCHANGE " + actor.PublicId +
// 							   " TO THE EXCHANGE LIST" ) }
// 	// Unidentified type of actor //
// 	} else if actor.Role != ADMIN_ROLE {
// 		return shim.Error( "ERROR: THE ACTOR ROLE " + actor.Role +
// 						   " IS NOT RECOGNISED." ) }

// 	// Store new registered Actor in Blockchain //
// 	actorBytes, _ := json.Marshal(actor)
// 	err16 := stub.PutState( IndexNetwork+actor.PublicId, actorBytes )
// 	if err16 != nil {
// 		return shim.Error( "ERROR REGISTERING ACTOR " + actor.PublicId +
// 						   " ON BLOCKCHAIN." + err16.Error() ) }

// 	// Prepare output with updates //
// 	users_update := make( map[string]Actor )
// 	users_update[actor.PublicId] = actor
// 	output.DID = encryption_DID
// 	output.UpdateUsers = users_update
// 	outputBytes, _ := json.Marshal( output )
// 	return shim.Success( outputBytes )
// }

// /* -------------------------------------------------------------------------------------------------
// getPrivacy:  this function returns a map containing the privacies of users from the different
//              companies. Args: array contaning:
// PublicId            string   // Public Id key of the user
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) getPrivacy( stub shim.ChaincodeStubInterface,
// 											    args []string) pb.Response {

// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: GETPRIVACY FUNCTION SHOULD BE CALLED " +
// 		                   "WITH ONE ARGUMENTS." ) }
// 	publicId := args[0]

// 	// Get state of user from Blockchain //
// 	userBytes, err := stub.GetState( IndexNetwork+publicId )
// 	if err != nil {
// 		return shim.Error( "ERROR: GETTING USER INFORMATION FOR " +
// 							publicId + ". " + err.Error() ) }

// 	return shim.Success( userBytes )
// }

// /* -------------------------------------------------------------------------------------------------
// modifyPrivacy: this function returns a map containing the privacies of users from the different
//                 companies. Args: array contaning a json with the following attributes:
// PublicId            string   // Public Id key of the user
// BusinessId          string   // Public Id key of the business
// Enabled             bool     // Field containing the privacy setting
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) modifyPrivacy( stub shim.ChaincodeStubInterface,
// 												   args []string) pb.Response {

// 	// Retrieve the privacy information for modification //
// 	privacyModifier := PrivacyModifier{}
// 	err1 := json.Unmarshal( []byte(args[0]), &privacyModifier )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: GETTING INPUT INFORMATION. " +
// 	                       err1.Error() )
// 	}

// 	// Get state of user from Blockchain //
// 	userBytes, err1 := stub.GetState( IndexNetwork+privacyModifier.PublicId )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: GETTING USER INFORMATION FOR " +
// 							privacyModifier.PublicId + ". " + err1.Error() )
// 	}

// 	// Change privacy settings of user //
// 	user := Actor{}
// 	json.Unmarshal( userBytes, &user )
// 	privacy := user.Privacy
// 	privacy[privacyModifier.BusinessId] = privacyModifier.Enabled
// 	user.Privacy = privacy

// 	// Store new state of user in Blockchain //
// 	userBytes2, _ := json.Marshal(user)
// 	err2 := stub.PutState( IndexNetwork+user.PublicId, userBytes2 )
// 	if err2 != nil {
// 		return shim.Error( "ERROR UPDATING USER " + user.PublicId +
// 	                       " ON BLOCKCHAIN." + err2.Error() )
// 	}
// 	return shim.Success( userBytes2 )
// }

// /* -------------------------------------------------------------------------------------------------
// getBusinessList:  this function returns the list of business registered in the PRIVI Blockchain. This
//                   function does not need to be called with any argument.
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) getBusinessList( stub shim.ChaincodeStubInterface,
// 													 args []string ) pb.Response {
// 	// Retrieve list of business from the Blockchain //
// 	businessList, err := stub.GetState( IndexBusinessList )
// 	if err != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE BUSINESS LIST" ) }
// 	return shim.Success( businessList )
// }

// /* -------------------------------------------------------------------------------------------------
// getUserList:  this function returns the list of users registered in the PRIVI Blockchain. This
//               function does not need to be called with any argument.
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) getUserList( stub shim.ChaincodeStubInterface,
// 											     args []string ) pb.Response {
// 	// Retrieve list of users from the Blockchain //
// 	userList, err := stub.GetState( IndexUserList )
// 	if err != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE USER LIST" )
// 	}

// 	return shim.Success( userList )
// }

// /* -------------------------------------------------------------------------------------------------
// getTIDList:  this function returns the list of TIDs generated for businesses. This
//              function does not need to be called with any argument.
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) getTIDList( stub shim.ChaincodeStubInterface,
// 												args []string ) pb.Response {
// 	// Retrieve list of TIDs from the Blockchain //
// 	TIDList, err := stub.GetState( IndexTargetEncryption + args[0] )
// 	if err != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE TID LIST" )
// 	}

// 	return shim.Success( TIDList )
// }

// /* -------------------------------------------------------------------------------------------------
// encryptData: this function is used to retrieve the DID of users to encrypt their data in the Cloud
//              Database. Args: array contaning:
// PublicId            string   // Public Id key of the user
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) encryptData( stub shim.ChaincodeStubInterface,
// 												 args []string ) pb.Response {
// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: ENCRYPTDATA TOKEN FUNCTION SHOULD BE CALLED " +
// 						   "WITH ONE ARGUMENTS." )
// 	}
// 	publicId := args[0]

// 	encryption_DID, err := stub.GetState( IndexEncryption+publicId )
// 	if err != nil {
// 		return shim.Error( "ERROR: RETRIEVING THE DID FOR USER " +
// 		                    publicId + ". " + err.Error() )
// 	}

// 	return shim.Success( encryption_DID )
// }

// /* -------------------------------------------------------------------------------------------------
// decryptData: this function is used to retrieve the PublicId key to decrypt the data stored in the
//              and encrypted in the Cloud Database. Args: array contaning:
// Encryption_DID           string   // Encrypting DID to decrypt
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) decryptData( stub shim.ChaincodeStubInterface,
// 												 args []string ) pb.Response {
// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: DECRYPTDATA TOKEN FUNCTION SHOULD BE CALLED " +
// 						   "WITH ONE ARGUMENT." )
// 	}
// 	encryption_DID := args[0]

// 	publicId, err := stub.GetState( IndexDecryption+encryption_DID)
// 	if err != nil {
// 		return shim.Error( "ERROR: DECRYPTING THE ID FOR DID " +
// 							encryption_DID + ". " + err.Error() )
// 	}
// 	return shim.Success( publicId )
// }

// /* -------------------------------------------------------------------------------------------------
// insightDiscovery: this function is used to retrieve the sublist of DIDs that have privacy enabled
//                   to the company. Args: array contaning:
// DID_list           string[]   // List of DIDs containing users to verify privacy
// ID_list            string[]   // List of IDs to filter from the DID_list
// Business_Id        string     // String of the business to get the ID
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) insightDiscovery( stub shim.ChaincodeStubInterface,
// 											          args []string ) pb.Response {
// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: INSIGHTDISCOVERY FUNCTION SHOULD BE " +
// 						   "CALLED WITH ONE ARGUMENT." )
// 	}
// 	insightDiscovery := InsightDiscovery{}
// 	json.Unmarshal( []byte(args[0]), &insightDiscovery )
// 	ID_DID_LIST := make( map[string]string )

// 	// Loop through list of DIDS and check privacies //
// 	for i, _ := range insightDiscovery.DID_list {
// 		DID := insightDiscovery.DID_list[i]
// 		// Decrypt user //
// 		publicIdBytes, err1 := stub.GetState( IndexDecryption+DID )
// 		if err1 != nil {
// 			return shim.Error( "ERROR: DECRYPTING THE ID FOR DID " +
// 								DID + ". " + err1.Error() )
// 		}
// 		publicId := string(publicIdBytes[:])
// 		// Get state of user from Blockchain //
// 		userBytes, err2 := stub.GetState( IndexNetwork+publicId )
// 		if err2 != nil {
// 			return shim.Error( "ERROR: GETTING USER INFORMATION. " +
// 			                   err2.Error() )
// 		}
// 		user := Actor{}
// 		json.Unmarshal( userBytes, &user)
// 		// Check if user has privacy enabled for business //
// 		if user.Privacy[ insightDiscovery.Business_Id ] {
// 			ID_DID_LIST[ publicId ] = DID
// 		}
// 	}

// 	// Loop through list of IDs and query form the list ID_DID_LIST //
// 	DID_filter := make( map[string]bool )
// 	DID_result := []string{}
// 	for _, ID := range insightDiscovery.ID_list {
// 		DID, inList := ID_DID_LIST[ID]
// 		if inList {
// 			DID_filter[DID] = true
// 			DID_result = append( DID_result, DID )
// 		}
// 	}

// 	fmt.Println( "-------->",DID_result )
// 	// Loop through TID List //
// 	TIDList := make( map[string]string )
// 	TIDListBytes, err := stub.GetState( IndexTargetEncryption + insightDiscovery.Business_Id  )
// 	json.Unmarshal( TIDListBytes, &TIDList )
// 	fmt.Println( "-------->", TIDList )
// 	if err == nil && len(TIDList) > 0 {
// 		fmt.Println("HEeeere")
// 		DID_result = []string{}
// 		for _, DID := range(TIDList) {
// 			inList, _ := DID_filter[DID]
// 			if inList {continue;}
// 			DID_result = append( DID_result, DID )
// 		}
// 	}

// 	result, _ := json.Marshal( DID_result )
// 	return shim.Success( result )
// }

// /* -------------------------------------------------------------------------------------------------
// insightPurchase: this function is called when a company purchases the insight of an user, the funds
//                  are distributed between users and providers. Args: array contaning:
// DID_list           		string[]   				// List of DIDs containing users to verify privacy
// Price              		float64    				// Price per user in PRIVI Data Token
// BusinessId              string    			    // Id of the business purchasing the insights
// InsightDistribution     map[string]float64      // Map containing distribution of price charged
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) insightPurchase( stub shim.ChaincodeStubInterface,
// 													 args []string ) pb.Response {
// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: INSIGHTPURCHASE FUNCTION SHOULD BE " +
// 						   "CALLED WITH ONE ARGUMENT." ) }

// 	insightPurchase := InsightPurchase{}
// 	json.Unmarshal( []byte(args[0]), &insightPurchase )
// 	price := insightPurchase.Price
// 	pct_user := insightPurchase.InsightDistribution["user"]
// 	total_charged := 0.

// 	// Retrieve business state to input TIDs //
// 	businessBytes, err1 := stub.GetState( IndexNetwork+insightPurchase.Business_Id )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: GETTING BUSINESS STATE FOR " +
// 							insightPurchase.Business_Id + ". " + err1.Error() )
// 	}
// 	business := Actor{}
// 	json.Unmarshal( businessBytes, &business )
// 	business_target_list := business.TargetTIDs

// 	// Retrieve TIDs encrypting list from State //
// 	TIDBytes, err2 := stub.GetState( IndexTargetEncryption +
// 									insightPurchase.Business_Id )
// 	if err2 != nil {
// 		return shim.Error( "ERROR: GETTING TID ENCRYPTING LIST. " +
// 		                    err2.Error() )
// 	}
// 	TID_DID_ENCRYPT := make( map[string]string )
// 	json.Unmarshal( TIDBytes, &TID_DID_ENCRYPT )

// 	// First add price transferred to users //
// 	transaction_list :=  []string{ "multitransfer" }
// 	for i, _ := range insightPurchase.DID_list {
// 		DID := insightPurchase.DID_list[i]
// 		// Decrypt user //
// 		publicIdBytes, err3 := stub.GetState( IndexDecryption+DID )
// 		if err3 != nil {
// 			return shim.Error( "ERROR: DECRYPTING THE ID FOR DID " +
// 								DID + ". " + err3.Error() )
// 		}
// 		publicId := string(publicIdBytes[:])
// 		// Check if user has privacy enabled for business //
// 		userBytes, err4 := stub.GetState( IndexNetwork+publicId )
// 		if err4 != nil {
// 			return shim.Error( "ERROR: GETTING USER INFORMATION. " +
// 			                   err4.Error() )
// 		}
// 		user := Actor{}
// 		json.Unmarshal( userBytes, &user)
// 		if !user.Privacy[ insightPurchase.Business_Id ] {continue}
// 		// Add transfer to transaction list //
// 		amount := price * pct_user
// 		userTransfer := Transfer{
// 			Type: "data_sale", Token: "PDT", From: insightPurchase.Business_Id,
// 			To: publicId, Amount: amount }
// 		userTransferBytes, _ := json.Marshal( userTransfer )
// 		transaction_list = append( transaction_list, string( userTransferBytes ) )
// 		// Encrypt DID in a Target encrypting Identifier TID //
// 		TID := xid.New().String() + xid.New().String() +
// 			   xid.New().String() + xid.New().String()
// 		TID_DID_ENCRYPT[TID] = DID
// 		business_target_list[TID] = true
// 		total_charged = total_charged + price
// 	}
// 	business.TargetTIDs = business_target_list

// 	// Second add price transferred to partners //
// 	for id_partner, pct_partner := range(insightPurchase.InsightDistribution) {
// 		if id_partner == "user" {continue}
// 		amount_partner := total_charged * pct_partner
// 		partnerTransfer := Transfer{
// 			Type: "data_sale", Token: "PDT", From: insightPurchase.Business_Id,
// 			To: id_partner, Amount: amount_partner }
// 		partnerTransferBytes, _ := json.Marshal( partnerTransfer )
// 		transaction_list = append( transaction_list, string( partnerTransferBytes ) )
// 	}

// 	// Perform the distribution of the charged amount //
// 	multiChainCodeArgs := ToChaincodeArgs( transaction_list )
// 	response := stub.InvokeChaincode( COIN_BALANCE_CHAINCODE, multiChainCodeArgs,
// 									  CHANNEL_NAME )
// 	if response.Status != shim.OK {
// 		return shim.Error( "ERROR DISTRIBUTING AMOUNT CHARGED FROM " +
// 							" DATA SALE. " + response.Message )
// 	}

// 	// Store Target Encrypting Identifiers (TIDs) on the Blockchain State //
// 	TIDBytes2, _ := json.Marshal( TID_DID_ENCRYPT )
// 	err5 := stub.PutState( IndexTargetEncryption + insightPurchase.Business_Id,
// 		                   TIDBytes2 )
// 	if err5 != nil {
// 		return shim.Error( "ERROR UPDATING TID ENCRYPTING LIST " +
// 	                       " ON BLOCKCHAIN." + err5.Error() )
// 	}

// 	// Update state of the businesss on Blockchain //
// 	businessBytes2, _ := json.Marshal(business)
// 	err6 := stub.PutState( IndexNetwork+insightPurchase.Business_Id, businessBytes2 )
// 	if err6 != nil {
// 		return shim.Error( "ERROR UPDATING BUSINESS " + insightPurchase.Business_Id +
// 	                       " ON BLOCKCHAIN." + err6.Error() )
// 	}

// 	output, _ := json.Marshal( business_target_list )
// 	return shim.Success( output )
// }

// /* -------------------------------------------------------------------------------------------------
// insightTarget: this function is called when a company wants to the users that he had
//                previously purchased to retrieve its identity. Args: array contaning:
// TID_list           		string[]   				// List of TIDs that the company wants to target
// Business_Id              string    			    // Id of the business purchasing the insights
// ------------------------------------------------------------------------------------------------- */

// func (t *DataProtocolSmartContract) insightTarget( stub shim.ChaincodeStubInterface,
// 													 args []string ) pb.Response {
// 	// Retrieve arguments of the input //
// 	if len(args) != 1 {
// 		return shim.Error( "ERROR: INSIGHTTARGET FUNCTION SHOULD BE " +
// 						   "CALLED WITH ONE ARGUMENT." )
// 	}
// 	insightTarget := InsightTarget{}
// 	json.Unmarshal( []byte(args[0]), &insightTarget )
// 	fmt.Println("INSGITH1 ------------------------------------> ", insightTarget)

// 	// Retrieve business state to input TIDs //
// 	businessBytes, err1 := stub.GetState( IndexNetwork+insightTarget.Business_Id )
// 	if err1 != nil {
// 		return shim.Error( "ERROR: GETTING BUSINESS STATE FOR " +
// 							insightTarget.Business_Id + ". " + err1.Error() )
// 	}
// 	business := Actor{}
// 	json.Unmarshal( businessBytes, &business )

// 	// Retrieve decryption of TIDs //
// 	TIDBytes, err2 := stub.GetState( IndexTargetEncryption +
// 									 insightTarget.Business_Id )
// 	if err2 != nil {
// 		return shim.Error( "ERROR GETTING THE TID ENCRYPTING LIST " +
// 	                       " ON BLOCKCHAIN." + err2.Error() )
// 	}
// 	TID_decryption := make( map[string]string )
// 	json.Unmarshal( TIDBytes, &TID_decryption )

// 	// Loop through all the TIDs and retrieve ID //
// 	ID_list := []string{}
// 	for _, TID := range(insightTarget.TID_list) {
// 		fmt.Println("INSGITH2 ------------------------------------> ", TID)
// 		fmt.Println("INSGITH3 ------------------------------------> ", business.TargetTIDs[TID] )
// 		fmt.Println("INSGITH4 ------------------------------------> ", TID_decryption[TID] )
// 		if business.TargetTIDs[TID] == false {continue}
// 		DID, inList := TID_decryption[TID]
// 		if inList == false {continue}
// 		// Decrypt user //
// 		publicIdBytes, err3 := stub.GetState( IndexDecryption+DID )
// 		if err3 != nil {
// 			return shim.Error( "ERROR: DECRYPTING THE ID FOR DID " +
// 								DID + ". " + err3.Error() )
// 		}
// 		publicId := string(publicIdBytes[:])
// 		// Check if user has privacy enabled for business //
// 		userBytes, err4 := stub.GetState( IndexNetwork+publicId )
// 		if err4 != nil {
// 			return shim.Error( "ERROR: GETTING USER INFORMATION. " +
// 			                   err4.Error() )
// 		}
// 		user := Actor{}
// 		json.Unmarshal( userBytes, &user)
// 		fmt.Println("INSGITH5 ------------------------------------> ", user )
// 		if !user.Privacy[ insightTarget.Business_Id ] {continue}
// 		// Add user Id to list and remove the TID //
// 		ID_list = append( ID_list, publicId )
// 		fmt.Println("INSGITH6 ------------------------------------> ", ID_list )
// 		delete(TID_decryption, TID)
// 		delete(business.TargetTIDs, TID)
// 	}

// 	// Update TIDs on the Blockchain State //
// 	TIDBytes2, _ := json.Marshal( TID_decryption )
// 	err5 := stub.PutState( IndexTargetEncryption + insightTarget.Business_Id,
// 		                   TIDBytes2 )
// 	if err5 != nil {
// 		return shim.Error( "ERROR UPDATING TID ENCRYPTING LIST " +
// 	                       " ON BLOCKCHAIN." + err5.Error() )
// 	}

// 	// Update state of the businesss on Blockchain //
// 	businessBytes2, _ := json.Marshal(business)
// 	err6 := stub.PutState( IndexNetwork+insightTarget.Business_Id, businessBytes2 )
// 	if err6 != nil {
// 		return shim.Error( "ERROR UPDATING BUSINESS " + insightTarget.Business_Id +
// 	                       " ON BLOCKCHAIN." + err6.Error() )
// 	}

// 	output, _ := json.Marshal( ID_list )
// 	return shim.Success( output )
// }

/* -------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------- */

func main() {
	err := shim.Start(&DataProtocolSmartContract{})
	if err != nil {
		fmt.Errorf("ERROR STARTING DATA PROTOCOL CHAINCODE: %s", err)
	}
}

/* -------------------------------------------------------------------------------------------------
------------------------------------------------------------------------------------------------- */
