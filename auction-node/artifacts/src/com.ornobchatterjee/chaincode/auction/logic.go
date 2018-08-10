package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"fmt"
	"time"
	"math/big"
	"reflect"
)

func (t *AuctionChaincode) getBidResult(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	org, _ := getMSPAttr(stub, MSP_ATTRIBUTE_ORG)
	if org != "Org2" {
		return shim.Error(fmt.Sprintf("Unauthorized user. Only Auction house users are allowed to invoke this function"))
	}

	t.Infof("[ getBidResult ] - Start")
	//pass the time string so as to always get a deterministic result
	currentTimeString := args[0]

	currentTime, err := time.Parse(time.RFC3339, currentTimeString)
	t.Infof("[ getBidResult ] - Current Time %v", currentTime.String())
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	//iterate over all assets that are not sold and having the bidEndTime in the past
	//query couchdb
	resultsIterator, err := stub.GetQueryResult(fmt.Sprint(QUERY_ALL_CLOSED_BIDS, currentTimeString))
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(getErrorString(err))
		}
		currentAssetKey := queryResponse.Key
		t.Infof("[ getBidResult ] - Current Asset Key for Bid Result %v", currentAssetKey)
		assetByte, err := stub.GetState(currentAssetKey)

		if err != nil {
			return shim.Error(getErrorString(err))
		}
		var assetObj Asset
		err = json.Unmarshal([]byte(assetByte), &assetObj)
		if err != nil {
			return shim.Error(getErrorString(err))
		}

		if (!assetObj.BidEnd.Before(currentTime)) || assetObj.IsSold {
			continue
		}
		t.Infof("[ getBidResult ] - Current Asset Id for Bid Result %v", assetObj.AssetId)

		t.declareWinnerForAsset(stub, &assetObj)
	}

	//filter the items which have completed the bidding cycle, over
	return shim.Success(nil)
}

func (t *AuctionChaincode) placeBid(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	org, _ := getMSPAttr(stub, MSP_ATTRIBUTE_ORG)
	if org != "Org1" {
		return shim.Error(fmt.Sprintf("Unauthorized user. Only general users are allowed to invoke this function"))
	}

	bidJsonString := args[0]
	var bidObj Bid
	var assetObj Asset
	var err error

	//check if user exists
	user, err := getUserByEmail(stub)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	err = json.Unmarshal([]byte(bidJsonString), &bidObj)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	// asset exists
	bidAssetOwnerEmail := bidObj.Asset.Owner.Email
	bidAssetId := bidObj.Asset.AssetId

	if len(bidAssetId) == 0 || len(bidAssetOwnerEmail) == 0 {
		return shim.Error(fmt.Sprintf("Asset ID and owner email is mandatory"))
	}
	bidAssetKey, err := getCompositeKey(stub, COMPOSITE_KEY_OWNER_ASSET, bidAssetOwnerEmail, string(bidAssetId))
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	foundAssetString, err := stub.GetState(bidAssetKey)
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
	//t.Infof(fmt.Sprint("Found Asset String %v", foundAssetString))
	// asset is not sold
	if assetObj.IsSold {
		return shim.Error(fmt.Sprintf("Asset : %v is already sold", bidAssetId))
	}

	if assetObj.Owner.Email == user.Email {
		return shim.Error(fmt.Sprintf("Asset : %v is already owned by bidding user", bidAssetId))
	}
	// bid amount is greater than or equal to the price of the asset
	if assetObj.Price.Cmp(bidObj.BidAmount) > 0 {
		return shim.Error(fmt.Sprintf("Asset : %v price is greater than bid price", bidAssetId))
	}

	//check if bidding time is within the limits
	bidStartTime := assetObj.BidStart
	bidEndTime := assetObj.BidEnd
	bidTime := bidObj.BidTime
	if bidStartTime != nil {
		t.Infof(fmt.Sprint("bidStartTime: %v", bidStartTime.String()))
	}

	if bidEndTime != nil {
		t.Infof(fmt.Sprint("bidEndTime %v", bidEndTime.String()))
	}

	if bidTime != nil {
		t.Infof(fmt.Sprint("bidTime: %v", bidTime.String()))
	}
	//TODO
	//if !(bidTime.After(*bidStartTime) && bidTime.Before(*bidEndTime)) {
	//	return shim.Error(fmt.Sprintf("Bid Time is not within the bounds of the start time and end time"))
	//}

	//check if user has sufficient balance to bid
	if user.Balance.Cmp(bidObj.BidAmount) < 0 {
		return shim.Error(fmt.Sprintf("User does not have sufficient amount to Bid"))
	}

	//place the bid
	bidCompositeKey, _ := getCompositeKey(stub, COMPOSITE_KEY_BID_ASSET_BIDDER, bidAssetId, user.Email)
	bidObj.DocType = reflect.TypeOf(bidObj).Name()
	bidObj.Asset = &assetObj

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

	org, _ := getMSPAttr(stub, MSP_ATTRIBUTE_ORG)
	if org != "Org1" {
		return shim.Error(fmt.Sprintf("Unauthorized user. Only general users are allowed to invoke this function"))
	}

	assetJson := args[0]
	currentTimeString := args[0]
	var assetObj Asset
	var err error

	currentTime, err := time.Parse(time.RFC3339, currentTimeString)

	err = json.Unmarshal([]byte(assetJson), &assetObj)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	//check if user exists
	user, err := getUserByEmail(stub)
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

	bidStartTime := assetObj.BidStart
	bidEndTime := assetObj.BidStart

	t.Infof(fmt.Sprint("Current Time %v", currentTime.String()))
	if bidEndTime.Before(*bidStartTime) {
		return shim.Error(fmt.Sprintf("Incorrect Bid Duration"))
	}

	//check if the bid duration is in the future
	if !(currentTime.Before(*bidStartTime) && currentTime.Before(*bidEndTime)) {
		return shim.Error(fmt.Sprintf("Bid Duration must be in the future"))
	}

	//set the reference of the
	assetObj.Owner = new(User)
	assetObj.Owner.Email = user.Email
	assetObj.IsSold = false
	assetObj.DocType = reflect.TypeOf(assetObj).Name()
	assetBytes, err := json.MarshalIndent(assetObj, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	if err = stub.PutState(ownerAssetCompositeKey, []byte(assetBytes)); err != nil {
		return shim.Error(getErrorString(err))
	}

	return shim.Success(nil)
}

func (t *AuctionChaincode) addUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	userString := args[0]
	var user User

	invokerEmail, err := getMSPAttr(stub, MSP_ATTRIBUTE_EMAIL)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	invokerEmailKey, _ := getCompositeKey(stub, USER_KEY, invokerEmail)
	invokerString, err := stub.GetState(invokerEmailKey)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	//user already exists
	if invokerString != nil {
		return shim.Error(fmt.Sprintf("User with email %v already exists", invokerEmail))
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
	user.Email = invokerEmail
	user.DocType = reflect.TypeOf(user).Name()
	userBytes, err := json.MarshalIndent(user, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return shim.Error(getErrorString(err))
	}

	if err = stub.PutState(invokerEmailKey, []byte(userBytes)); err != nil {
		return shim.Error(getErrorString(err))
	}

	return shim.Success(nil)
}

/**
Transfer the asset from original owner to new owner with the provided email
 */
func (t *AuctionChaincode) transferAsset(stub shim.ChaincodeStubInterface, assetObj *Asset, newOwnerEmail string) (error) {
	oldAssetKey, _ := getCompositeKey(stub, COMPOSITE_KEY_OWNER_ASSET, assetObj.Owner.Email, assetObj.AssetId)
	err := stub.DelState(oldAssetKey)
	if err != nil {
		return err
	}
	assetObj.Owner.Email = newOwnerEmail
	assetObj.IsSold = true
	newAssetKey, _ := getCompositeKey(stub, COMPOSITE_KEY_OWNER_ASSET, newOwnerEmail, assetObj.AssetId)
	newAssetBytes, err := json.MarshalIndent(assetObj, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return err
	}
	if err = stub.PutState(newAssetKey, []byte(newAssetBytes)); err != nil {
		return err
	}
	return nil
}

func (t *AuctionChaincode) declareWinnerForAsset(stub shim.ChaincodeStubInterface, assetObj *Asset) (error) {
	t.Infof("[ declareWinnerForAsset ] - start for asset id %v", assetObj.AssetId)

	assetId := assetObj.AssetId
	availableBidsIterator, err := stub.GetStateByPartialCompositeKey(COMPOSITE_KEY_BID_ASSET_BIDDER, []string{assetId})
	if err != nil {
		return err
	}

	var maxBid = new(big.Rat)
	maxBid.SetString("0")
	var maxBidderEmail string
	hasBids := availableBidsIterator.HasNext()

	t.Infof("[ declareWinnerForAsset ] - hasBids %v", hasBids)

	defer availableBidsIterator.Close()
	for availableBidsIterator.HasNext() {
		var currBidObj Bid
		responseRange, err := availableBidsIterator.Next()
		if err != nil {
			return err
		}
		currentBidValBytes := responseRange.Value
		currentBidKey := responseRange.Key
		t.Infof("[ declareWinnerForAsset ] - currentBidKey %v", currentBidKey)

		_, currentBidKeyParts, _ := stub.SplitCompositeKey(currentBidKey)
		currentBidderEmail := currentBidKeyParts[1]

		t.Infof("[ declareWinnerForAsset ] - currentBidderEmail %v", currentBidderEmail)

		err = json.Unmarshal([]byte(currentBidValBytes), &currBidObj)
		if err != nil {
			return err
		}
		if currBidObj.BidAmount.Cmp(maxBid) > 0 {
			maxBid = currBidObj.BidAmount
			maxBidderEmail = currentBidderEmail
		}
	}

	if hasBids {
		t.Infof("[ declareWinnerForAsset ] - maxBidderEmail %v", maxBidderEmail)
		t.Infof("[ declareWinnerForAsset ] - maxBid %v", maxBid.String())

		user, err := getUserByEmail(stub, maxBidderEmail)
		if err != nil {
			return err
		}
		//deduct the balance
		user.Balance.Sub(user.Balance, maxBid)
		t.Infof("[ declareWinnerForAsset ] - userId %v has Balance %v ", user.UserId, user.Balance)

		maxBidderEmailKey, _ := getCompositeKey(stub, USER_KEY, maxBidderEmail)

		userBytes, err := json.MarshalIndent(user, JSON_PREFIX, JSON_INDENT)
		if err != nil {
			return err
		}
		if err = stub.PutState(maxBidderEmailKey, []byte(userBytes)); err != nil {
			return err
		}

		//add the balance to the owner of the asset
		originalOwnerEmail := assetObj.Owner.Email
		originalOwner, err := getUserByEmail(stub, originalOwnerEmail)
		if err != nil {
			return err
		}
		originalOwner.Balance.Add(originalOwner.Balance, maxBid)
		t.Infof("[ declareWinnerForAsset ] - userId %v has Balance %v ", originalOwner.UserId, originalOwner.Balance)

		originalOwnerEmailKey, err := getCompositeKey(stub, USER_KEY, originalOwnerEmail)
		originalOwnerBytes, err := json.MarshalIndent(originalOwner, JSON_PREFIX, JSON_INDENT)
		if err != nil {
			return err
		}
		if err = stub.PutState(originalOwnerEmailKey, []byte(originalOwnerBytes)); err != nil {
			return err
		}
		t.transferAsset(stub, assetObj, maxBidderEmail)
	}
	return nil
}

func (t *AuctionChaincode) getAssetsForUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	mode := args[0]
	user, err := getUserByEmail(stub);
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	var availableAssetsIterator shim.StateQueryIteratorInterface
	if mode == "all" {
		availableAssetsIterator, err = stub.GetStateByPartialCompositeKey(COMPOSITE_KEY_OWNER_ASSET, []string{})
	}else {
		availableAssetsIterator, err = stub.GetStateByPartialCompositeKey(COMPOSITE_KEY_OWNER_ASSET, []string{user.Email})
	}

	if err != nil {
		return shim.Error(getErrorString(err))
	}

	var assets = make([]Asset,0);

	defer availableAssetsIterator.Close()
	for availableAssetsIterator.HasNext() {
		var currAssetObj Asset
		responseRange, err := availableAssetsIterator.Next()
		if err != nil {
			return shim.Error(getErrorString(err))
		}
		currentAssetValBytes := responseRange.Value
		currentBidKey := responseRange.Key

		_, currentBidKeyParts, _ := stub.SplitCompositeKey(currentBidKey)
		currentBidderEmail := currentBidKeyParts[1]

		t.Infof("[ declareWinnerForAsset ] - currentBidderEmail %v", currentBidderEmail)

		err = json.Unmarshal([]byte(currentAssetValBytes), &currAssetObj)
		if err != nil {
			return shim.Error(getErrorString(err))
		}
		assets = append(assets, currAssetObj)
	}
	assetsBytes,err := json.MarshalIndent(assets, JSON_PREFIX, JSON_INDENT)
	return shim.Success(assetsBytes);
}

func (t *AuctionChaincode) getUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var user *User;
	var err error;
	if len(args) == 1 {
		user, err = getUserByEmail(stub, args[0]);
	}else {
		user, err = getUserByEmail(stub);
	}
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	userBytes, err := json.MarshalIndent(&user, JSON_PREFIX, JSON_INDENT)
	if err != nil {
		return shim.Error(getErrorString(err))
	}
	return shim.Success(userBytes)
}
