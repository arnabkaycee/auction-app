package main

import (
	"math/big"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type User struct {
	UserId       string   `json:"userId,omitempty"`
	Email        string   `json:"email,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Balance      *big.Rat `json:"balance,omitempty"`
	Organization string   `json:"org,omitempty"`
	DocType      string   `json:"docType,omitempty"`
}

type Asset struct {
	AssetId     string     `json:"assetId,omitempty"`
	Owner       *User      `json:"owner,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Price       *big.Rat   `json:"price,omitempty"`
	BidStart    *time.Time `json:"bidStart,omitempty"`
	BidEnd      *time.Time `json:"bidEnd,omitempty"`
	IsSold      bool       `json:"isSold,omitempty"`
	DocType     string     `json:"docType,omitempty"`
}

type Bid struct {
	BidId     string     `json:"bidId,omitempty"`
	Asset     *Asset     `json:"asset,omitempty"`
	BidAmount *big.Rat   `json:"bidAmount,omitempty"`
	BidTime   *time.Time `json:"bidTime,omitempty"`
	DocType   string     `json:"docType,omitempty"`
}

//func toJSON(anyStruct interface{}) ([]byte) {
//	bytes, _ := json.MarshalIndent(anyStruct, JSON_PREFIX, JSON_INDENT)
//	return bytes
//}
//
//func fromJSON(jsonBytes []byte, anyInterface interface{}) (interface{})  {
//	instance := reflect.New(reflect.TypeOf(anyInterface));
//	_ = json.Unmarshal(jsonBytes, &instance)
//	return instance;
//}

type AuctionChaincode struct {
	*shim.ChaincodeLogger
}
