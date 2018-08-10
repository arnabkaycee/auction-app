package main

import (
	"math/big"
	"time"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type User struct {
	UserId  string  `json:"userId"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Balance *big.Rat `json:"balance"`
}

type Asset struct {
	AssetId     string   `json:"assetId,omitempty"`
	Owner       *User    `json:"owner,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Price       *big.Rat  `json:"price,omitempty"`
	BidStart    JSONTime `json:"bidStart,omitempty"`
	BidEnd      JSONTime `json:"bidEnd,omitempty"`
	IsSold      bool     `json:"isSold,omitempty"`
}

type Bid struct {
	BidId     string   `json:"bidId,omitempty"`
	AssetId   string   `json:"assetId,omitempty"`
	UserId    string   `json:"userId,omitempty"`
	BidAmount *big.Rat  `json:"bidAmount,omitempty"`
	BidTime   JSONTime `json:"bidStart,omitempty"`
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("Mon Jan _2"))
	return []byte(stamp), nil
}

type AuctionChaincode struct {
	*shim.ChaincodeLogger
}
