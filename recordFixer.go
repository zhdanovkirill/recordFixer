package main

/* Imports
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"strings"
	"reflect"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

const NOT_FOUND_ERROR string = "NOT_FOUND_ERROR"
const ALREADY_EXIST_ERROR string = "ALREADY_EXIST_ERROR"

// Define the Smart Contract structure
type SmartContract struct {
}

// type Document interface {
//  DocumentId() string
// }
type Document struct {
	UserName string `json:"userName"`
	Subject  string `json:"subject"`
	SendTime string `json:"sendTime"`
}

/*
 * The Init method is called when the Smart Contract "" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */// Init is called during chaincode instantiation to initialize any data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {

}
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "recordFixer"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "getDocument" {
		return s.getDocument(APIstub, args)
	} else if function == "registerDocument" {
		return s.registerDocument(APIstub, args)
	} else if function == "updateDocument" {
		return s.updateDocument(APIstub, args)
	} else if function == "revokeDocument" {
		return s.revokeDocument(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) getDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 || len(args) > 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1 or 2")
	}

	documentAsBytes, err := APIstub.GetState(args[0])
	if len(documentAsBytes) == 0 {
		//return shim.Success(nil);
		return shim.Error(NOT_FOUND_ERROR)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// buffer is a JSON array containing historic values for recordFixer
	var buffer bytes.Buffer
	if len(args) == 2 && args[1] == "full" {

		resultsIterator, err := APIstub.GetHistoryForKey(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		defer resultsIterator.Close()

		buffer.WriteString(", \"History\": ")
		buffer.WriteString("[")

		bArrayMemberAlreadyWritten := false
		for resultsIterator.HasNext() {
			response, err := resultsIterator.Next()
			if err != nil {
				return shim.Error(err.Error())
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(response.TxId)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Value\":")
			if response.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(response.Value))
			}

			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
			buffer.WriteString("\"")

			buffer.WriteString(", \"IsDelete\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatBool(response.IsDelete))
			buffer.WriteString("\"")

			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
		buffer.WriteString("]")
		buffer.WriteString("}")

		documentAsBytes = documentAsBytes[0 : len(documentAsBytes)-1]
		documentAsBytes = append(documentAsBytes, buffer.Bytes()...)
	}
	//add id
	buffer.Reset()
	buffer.WriteString(", \"id\":\"")
	buffer.WriteString(args[0])
	buffer.WriteString("\"}")
	documentAsBytes = documentAsBytes[0 : len(documentAsBytes)-1]
	documentAsBytes = append(documentAsBytes, buffer.Bytes()...)

	return shim.Success(documentAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) registerDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	documentAsBytes, err := APIstub.GetState(args[0])

	if len(documentAsBytes) != 0 {
		return shim.Error(ALREADY_EXIST_ERROR)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	var document = Document{Name: args[1], IssuerId: args[2], IssuedAt: args[3], Description: args[4], ExpiresAt: args[5], IssuedTo: args[6], Revoked: strings.Compare(args[7], "true") == 0, RevokeReason: args[8], DocumentData: args[9], RegisteredAt: args[10]}
	documentAsBytes, err = json.Marshal(document)
	APIstub.PutState(args[0], documentAsBytes)

	return shim.Success([]byte(args[0]))
}

func (s *SmartContract) revokeDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	documentAsBytes, err := APIstub.GetState(args[0])
	if len(documentAsBytes) == 0 {
		return shim.Error(NOT_FOUND_ERROR)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	doc := Document{}

	json.Unmarshal(documentAsBytes, &doc)
	doc.Revoked = true
	doc.RevokeReason = args[1]

	documentAsBytes, _ = json.Marshal(doc)
	APIstub.PutState(args[0], documentAsBytes)

	return shim.Success([]byte(args[0]))
}

func (s *SmartContract) updateDocument(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	documentAsBytes, err := APIstub.GetState(args[0])
	if len(documentAsBytes) == 0 {
		return shim.Error(NOT_FOUND_ERROR)
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	doc := Document{}
	json.Unmarshal(documentAsBytes, &doc)

	d := map[string]interface{}{}
	err = json.Unmarshal([]byte(args[1]), &d)
	if err != nil {
		return shim.Error("Error convert args[1] to JSON")
	}

	//copy all property from m to document (replace existing)
	obj := &doc
	ro := reflect.ValueOf(obj).Elem()
	typeOfT := ro.Type()
	for i := 0; i < ro.NumField(); i++ {
		for j, f := range d {
			if typeOfT.Field(i).Tag.Get("json") == j {
				fl := ro.FieldByName(typeOfT.Field(i).Name)
				switch fl.Kind() {
				case reflect.Bool:
					fl.SetBool(f.(bool))
				case reflect.Int, reflect.Int64:
					c, _ := f.(float64)
					fl.SetInt(int64(c))
				case reflect.String:
					fl.SetString(f.(string))
				}
			}
		}
	}

	documentAsBytes, err = json.Marshal(doc)
	APIstub.PutState(args[0], documentAsBytes)

	return shim.Success([]byte(args[0]))
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
