#!/bin/bash

AH_TOKEN=`cat ~/auction/user_token.json | jq '.["auctionhouse"]' | tr -d '\"'`


while true; do
    CURRENT_TIME=`node -e 'console.log(new Date())'`
	echo "Running Bid Result Batch at "$CURRENT_TIME
    curl -s -X POST \
   http://localhost:4000/channels/mychannel/chaincodes/mycc \
   -H "authorization: Bearer $AH_TOKEN" \
   -H "content-type: application/json" \
   -d '{
 	"peers": ["peer0.org1.example.com","peer1.org1.example.com","peer0.org2.example.com",
        "peer1.org2.example.com"],
 	"fcn":"getBidResult",
 	"args":["'$CURRENT_TIME'"]
    }'
    echo "Run completed"
	sleep 5
done    
