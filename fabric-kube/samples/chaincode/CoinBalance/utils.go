package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
)

/* -------------------------------------------------------------------------------------------------
-------------------------------------------------------------------------------------------------*/

func parsePEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}

// Extracts CN from an x509 certificate //
func CNFromX509(certPEM string) (string, error) {
	cert, err := parsePEM(certPEM)
	if err != nil {
		return "", errors.New("Failed to parse certificate: " + err.Error())
	}

	return cert.Subject.CommonName, nil
}

// Extracts CN from caller of a chaincode function //
func CallerCN(stub shim.ChaincodeStubInterface) (string, error) {
	data, _ := stub.GetCreator()
	fmt.Println(string(data), "data")
	serializedId := msp.SerializedIdentity{}
	err := proto.Unmarshal(data, &serializedId)
	if err != nil {
		return "", errors.New("Could not unmarshal creator")
	}
	cn, err := CNFromX509(string(serializedId.IdBytes))
	if err != nil {
		return "", err
	}
	fmt.Println(cn, "cn")
	return cn, nil
}

/*--------------------------------------------------
	Convert inputs to chaincodearguments
----------------------------------------------------*/

func ToChaincodeArgs(args []string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

/* -------------------------------------------------------------------------------------------------
These are utility functions
 -------------------------------------------------------------------------------------------------*/

func getTimeNow() string {
	var formatedTime string
	t := time.Now()
	formatedTime = t.Format(time.RFC1123)
	return formatedTime
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func saveSubstraction(main float64, amount float64) (float64, error) {
	main -= amount
	if main < 0. {
		return main, errors.New("ERROR: INSUFFICIENT FUNDS ON BALANCE")
	}
	return main, nil
}

func saveAddition(main float64, amount float64) (float64, error) {
	if main+amount < main {
		return main, errors.New("ERROR: OVERFLOW ON RECEIVER BALANCE")
	}
	main += amount
	return main, nil
}

func checkRange(number float64, lowerBound float64, upperBound float64) bool {
	if number > upperBound || number < lowerBound {
		return false
	}
	return true
}

func toStringMethod(object interface{}) string {
	objectBytes, _ := json.Marshal(object)
	return string(objectBytes)
}

func removeList(list []string, element string) []string {
	var elementIdx int
	for i, b := range list {
		if b == element {
			elementIdx = i
		}
	}
	list[elementIdx] = list[len(list)-1] // Copy last element to index i.
	list[len(list)-1] = ""               // Erase last element (write zero value).
	list = list[:len(list)-1]
	return list
}

func mergeMaps(map1 map[string]Balance,
	map2 map[string]Balance) map[string]Balance {
	for k, v := range map2 {
		map1[k] = v
	}
	return map1
}

/* -------------------------------------------------------------------------------------------------
-------------------------------------------------------------------------------------------------*/
