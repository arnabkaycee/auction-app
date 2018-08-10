package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"fmt"
	"time"
	"math/big"
)

func (t *AuctionChaincode) getBidResult(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//var i int32;
	//resultsIterator, err := stub.GetStateByPartialCompositeKey(COMPOSITE_KEY_BID_ASSET_BIDDER, []string{})
	//if err != nil {
	//	return shim.Error(getErrorString(err))
	//}
	//defer resultsIterator.Close()
	//for i = 0; resultsIterator.HasNext(); i++ {
	//
	//	//get the list of bids in
	//	responseRange, err := resultsIterator.Next()
	//
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//}
	return shim.Success(nil)
}

func (t *AuctionChaincode) placeBid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	bidJsonString := args[0]
	var bidObj Bid
	var assetObj Asset
	var err error

	//check if user exists
	user, err := getUser(stub);
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	err = json.Unmarshal([]byte(bidJsonString), &bidObj)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	// asset exists
	bidAssetId := bidObj.AssetId
	foundAssetString, err := stub.GetState(bidAssetId)
	if foundAssetString == nil {
		return shim.Error(fmt.Sprintf("Asset : %v is not found", bidAssetId))
	}
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	//unmarshal the bid
	err = json.Unmarshal([]byte(foundAssetString), &assetObj)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	// asset is not sold
	if assetObj.IsSold {
		return shim.Error(fmt.Sprintf("Asset : %v is already sold", bidAssetId))
	}

	if assetObj.Owner.UserId == user.UserId {
		return shim.Error(fmt.Sprintf("Asset : %v is already owned by bidding user", bidAssetId))
	}
	// bid amount is greater than or equal to the price of the asset
	if assetObj.Price.Cmp(bidObj.BidAmount) > 0 {
		return shim.Error(fmt.Sprintf("Asset : %v price is greater than bid price", bidAssetId))
	}

	//check if bidding time is within the limits
	bidStartTime := time.Time(assetObj.BidStart)
	bidEndTime := time.Time(assetObj.BidStart)
	bidTime := time.Time(bidObj.BidTime)
	if !(bidTime.After(bidStartTime) && bidTime.Before(bidEndTime)) {
		return shim.Error(fmt.Sprintf("Bid Time is not within the bounds of the start time and end time"))
	}

	//check if user has sufficient balance to bid
	if user.Balance.Cmp(bidObj.BidAmount) < 0 {
		return shim.Error(fmt.Sprintf("User does not have sufficient amount to Bid"))
	}

	//place the bid
	bidCompositeKey, _ := getCompositeKey(stub, COMPOSITE_KEY_BID_ASSET_BIDDER, bidAssetId, user.UserId)
	bidBytes, err := json.MarshalIndent(bidObj, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	if err = stub.PutState(bidCompositeKey, []byte(bidBytes)); err != nil {
		return shim.Error(getErrorString(err))
	}
	return shim.Success(nil)
}

func (t *AuctionChaincode) addAssetForBid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	assetJson := args[0]
	var assetObj Asset
	var err error

	err = json.Unmarshal([]byte(assetJson), &assetObj)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	//check if user exists
	user, err := getUser(stub);
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	//check if asset already exists with the given id for the same owner
	ownerAssetCompositeKey, _ := getCompositeKey(stub, COMPOSITE_KEY_OWNER_ASSET, user.Email, assetObj.AssetId)
	foundAssetString, err := stub.GetState(ownerAssetCompositeKey)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	if foundAssetString != nil {
		return shim.Error(fmt.Sprintf("Asset with the same Id already belongs to the owner"))
	}
	var ZERO = new(big.Rat)
	ZERO.SetString("0")
	//check if the price of the item is greater than zero
	if assetObj.Price.Cmp(ZERO) <= 0 {
		return shim.Error(fmt.Sprintf("Asset : %v price cannot be zero or less than zero", assetObj.AssetId))
	}
	bidStartTime := time.Time(assetObj.BidStart)
	bidEndTime := time.Time(assetObj.BidStart)
	timeNow := time.Now()

	if bidEndTime.Before(bidStartTime) {
		return shim.Error(fmt.Sprintf("Incorrect Bid Duration"))
	}

	//check if the bid duration is in the future
	if !(timeNow.Before(bidStartTime) && timeNow.Before(bidEndTime)) {
		return shim.Error(fmt.Sprintf("Bid Duration must be in the future"))
	}

	return shim.Success(nil)
}

func (t *AuctionChaincode) addUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	userString := args[0];
	var user User

	invokerEmail, err := getMSPAttr(stub, MSP_ATTRIBUTE_EMAIL)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	invokerEmailKey, _ := getCompositeKey(stub, USER_KEY, USER_KEY, invokerEmail)
	invokerString, err := stub.GetState(invokerEmailKey)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	//user already exists
	if invokerString != nil {
		return shim.Error(fmt.Sprintf("User with email %v already exists", invokerEmail));
	}

	err = json.Unmarshal([]byte(userString), &user)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	if len(user.UserId) == 0 {
		return shim.Error(fmt.Sprintf("User Id is mandatory"))
	}

	if user.Balance == nil {
		return shim.Error(fmt.Sprintf("User Balance is mandatory"))
	}
	user.Email = invokerEmail;
	userBytes, err := json.MarshalIndent(user, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	if err = stub.PutState(invokerEmailKey, []byte(userBytes)); err != nil {
		return shim.Error(getErrorString(err))
	}

	return shim.Success(nil)
}
