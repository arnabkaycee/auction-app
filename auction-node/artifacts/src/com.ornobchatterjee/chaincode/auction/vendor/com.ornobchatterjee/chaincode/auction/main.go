package main

import (
	//"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
)


func (t *AuctionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *AuctionChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	t.Debugf("Invoke: function=%q args=%v", function, args)

	type invokeFunc func(stub shim.ChaincodeStubInterface, args []string) pb.Response
	var invokeFunctions = map[string]struct {
		function invokeFunc
		nrArgs   int
	}{
		"addUser":        {t.addUser, 1},
		"addAssetForBid": {t.addAssetForBid, 1},
		"placeBid":       {t.placeBid, 1},
		"getBidResult":   {t.getBidResult, 1},
	}

	if fn, ok := invokeFunctions[function]; !ok {
		return shim.Error("unknown invoke function")
	} else if fn.nrArgs != len(args) {
		return shim.Error(fmt.Sprintf("incorrect number of arguments; expected %v, found %v ", fn.nrArgs, len(args)))
	} else {
		return fn.function(stub, args)
	}

}

func New(log *shim.ChaincodeLogger) *AuctionChaincode {
	return &AuctionChaincode{log}
}

func main() {
	log := shim.NewLogger(name)
	log.Debug("Starting chaincode...")
	if err := shim.Start(New(log)); err != nil {
		log.Criticalf("Error starting chaincode: %s\n", err)
	}
}

