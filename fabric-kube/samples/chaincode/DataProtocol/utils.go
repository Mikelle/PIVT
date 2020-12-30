package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
)

func parsePEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}

// extracts CN from an x509 certificate
func CNFromX509(certPEM string) (string, error) {
	cert, err := parsePEM(certPEM)
	if err != nil {
		return "", errors.New("Failed to parse certificate: " + err.Error())
	}

	return cert.Subject.CommonName, nil
}

// extracts CN from caller of a chaincode function
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

func toStringMethod(object interface{}) string {
	objectBytes, _ := json.Marshal(object)
	return string(objectBytes)
}
