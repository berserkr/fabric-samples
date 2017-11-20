/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_cc0")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// The address is a traditional bc address
// The email is for identification purpose, like an account
// The type can be: facial, password, pin, key, etc...
// The payload depennds on the type, where it can be a pub key, or bloom filters
// Bloom filters encoded are as follows: 'order_number/bitstream'
type Factor struct {
	Address string   `json:"address"`
	Email   string   `json:"email"`
	Type    string   `json:"type"`
	Payload []string `json:"payload"`
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Init ###########")

	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if function == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if function == "move" {
		// Deletes an entity from its state
		return t.move(stub, args)
	}
	if function == "write" {
		// Writes onto the state
		return t.write(stub, args)
	}
	if function == "read" {
		// Reads from teh state
		return t.read(stub, args)
	}
	if function == "storeFactor" {
		// Reads from teh state
		return t.storeFactor(stub, args)
	}
	if function == "getFactor" {
		// Reads from teh state
		return t.getFactor(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", args[0]))
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	logger.Infof("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var key string   // The key
	var value string // The data value
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3, function followed by 1 key and 1 value")
	}

	key = args[0]
	value = args[1]

	logger.Infof("Writting: key = %s, value = %s\n", key, value)

	// Write the state back to the ledger
	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var key string   // The key
	var value string // The return value
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2, function followed by 1 key")
	}

	key = args[0]

	logger.Infof("Reading: key = %s\n", key)

	valbytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if valbytes == nil {
		return shim.Error("Entity not found")
	}
	value = string(valbytes)

	logger.Infof(" Response:%s\n", value)
	return shim.Success(valbytes)
}

// This method will write a factor to the world state...
func (t *SimpleChaincode) storeFactor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var factor Factor // A factor...

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2, function followed by the factor data")
	}

	// get the bytes from the argument
	factorAsBytes := []byte(args[0])
	factor = Factor{}
	json.Unmarshal(factorAsBytes, &factor)

	fmt.Println("Added", factor)

	logger.Infof("Adding factor with address %s\n", factor.Address)

	// Write the state back to the ledger
	err := stub.PutState(factor.Address, factorAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// This method will return the factor from the state
func (t *SimpleChaincode) getFactor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var address string // The factor address
	var factor Factor

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2, function followed by 1 address")
	}

	address = args[0]

	logger.Infof("Reading factor with address = %s\n", address)

	factorAsBytes, err := stub.GetState(address)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if factorAsBytes == nil {
		return shim.Error("Entity not found")
	}

	factor = Factor{}

	json.Unmarshal(factorAsBytes, &factor)
	fmt.Println("Returning", factor)

	logger.Infof("Returning factor for %s with address %s\n", factor.Email, factor.Address)
	return shim.Success(factorAsBytes)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Infof("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
