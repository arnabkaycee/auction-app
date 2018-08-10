package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"runtime"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"encoding/json"
)

var Logger = shim.NewLogger("auction_cc")

const (
	name                = "auction-chaincode"
	MSP_ATTRIBUTE_EMAIL = "email"
	MSP_ATTRIBUTE_ORG   = "org"
)

const (
	JSON_PREFIX = ""
	JSON_INDENT = "    "
)

const (
	COMPOSITE_KEY_OWNER_ASSET      = "owner~asset"
	COMPOSITE_KEY_BID_ASSET_BIDDER = "asset~bidder"
	USER_KEY                       = "user~email"
)

const (
	QUERY_ALL_CLOSED_BIDS = "{\"selector\":{\"docType\":\"Asset\",\"bidEnd\":{\"$gt\":\"%v\"}}}"
)

func getCompositeKey(stub shim.ChaincodeStubInterface, keyConstant string, keys ...string) (string, error) {
	key, err := stub.CreateCompositeKey(keyConstant, keys)
	if err != nil {
		return "", errors.New("Error in creating composite Key : " + err.Error())
	}
	return key, nil
}

func getErrorString(err error) string {
	_, fn, line, _ := runtime.Caller(1)
	errStr := fmt.Sprintf("[error] %s:%d %v", fn, line, err)
	return errStr
}

func getMSPAttr(stub shim.ChaincodeStubInterface, attribute string) (string, error) {
	val, _, err := cid.GetAttributeValue(stub, attribute);
	if err != nil {
		return "", errors.New("Error in retrieving MSP attribute value : " + err.Error())
	}
	return val, nil;
}

func getUserByEmail(stub shim.ChaincodeStubInterface, userEmail ...string) (*User, error) {

	var user User
	invokerEmail, err := getMSPAttr(stub, MSP_ATTRIBUTE_EMAIL)
	if err != nil {
		return nil, err
	}

	queryEmail := invokerEmail;
	if len(userEmail) > 0 {
		queryEmail = userEmail[0];
	}

	emailKey, _ := getCompositeKey(stub, USER_KEY, queryEmail)
	invokerString, err := stub.GetState(emailKey)
	if err != nil {
		return nil, err
	}
	if invokerString == nil {
		return nil, errors.New("user not registered")
	}
	err = json.Unmarshal([]byte(invokerString), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil;
}
