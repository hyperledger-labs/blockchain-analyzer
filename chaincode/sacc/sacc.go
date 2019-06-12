/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (s *SimpleAsset) Init(stub shim.ChaincodeStubInterface) sc.Response {

	/*v := uint32(500)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)*/

	value := int64(10000000)
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, value)

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState("key0", buf)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", "key0"))
	}
	return shim.Success(nil)
}

func (s *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	fn, args := stub.GetFunctionAndParameters()

	if fn == "set" {
		return s.setData(stub, args)
	} else if fn == "get" {
		return s.queryData(stub, args)
	}

	return shim.Error("Invalid invoke function name")
}

func (s *SimpleAsset) setData(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	v, err0 := strconv.ParseInt(args[1], 10, 64)

	if err0 != nil {
		return shim.Error(fmt.Sprintf("Failed to parse args[1]: %s", args[1]))
	}

	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, v)

	// We store the key and the value on the ledger
	err1 := stub.PutState(args[0], buf)
	if err1 != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

func (s *SimpleAsset) queryData(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	dataAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Query failed: %s", args[0]))
	}
	return shim.Success(dataAsBytes)
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
