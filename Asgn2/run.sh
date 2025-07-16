#!/bin/bash

OPTION=$1
TESTNET="../test-network/network.sh"
CHAINCODE="../Asgn2/chaincode"
WALLET="./app/wallet"

if [ "$OPTION" == "launch" ]
then
    "$TESTNET" down
    rm -rf "$WALLET"
    "$TESTNET" up -ca
    "$TESTNET" createChannel
    "$TESTNET" deployCC -ccn chaincode -ccp "$CHAINCODE" -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
elif [ "$OPTION" == "kill" ]
then
    "$TESTNET" down
    rm -rf "$WALLET"
else
    echo "Help:
            ./run.sh launch
            ./run.sh kill"
fi
