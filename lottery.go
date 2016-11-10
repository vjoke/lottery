/**
 * @author: hsiung
 * @date: 2016/11/3
 * @desc: a simple lottery
 */

 package main
 
 import (
 	"time"
 	"errors"
 	"fmt"
	"strconv"
	"math/big"
	"math"
 	"crypto/md5"
 	"crypto/rand"
 	"encoding/base64"
 	"encoding/hex"
 	"encoding/json"
 	"io"
 	"github.com/hyperledger/fabric/core/chaincode/shim"
 )

 type SimpleChaincode struct {

 }

 // FIXME: global vars
 var OutletNo int = 1
 var PlayerNo int = 1
 var TicketNo int = 1

 // actor category
 const (
 	ISSUER = 1
 	BRANCH = 2
 	OUTLET = 3
 	PLAYER = 4
 )

 const (
 	ISSUER_ID = 0
 	LOTTERY_ID = 0
 	TICKET_PERIOD = 100
 )

 const (
 	ISSUER_PREFIX = "issuer-"
 	LOTTERY_PREFIX= "lottery-"
 	OUTLET_PREFIX = "outlet-"
 	PLAYER_PREFIX = "player-"
 	TICKET_PREFIX = "ticket-"
 )

 // ticket state
 const (
 	INVALID = 0
 	VALID 	= 1
 	MISSED 	= 2
 	WON 	= 3
 ) 

 type Issuer struct {
 	ID int
 	Name string
 }

 // type Branch struct {
 // 	ID int
 // 	Name string 		
 // }

 type Lottery struct {
 	ID int
 	IssuerId int
 	Round int
 	TicketAddress []int
 }

 type Outlet struct {
 	ID int
 	Name string
 	Count int
 }

 type Player struct {
 	ID int
 	Name string
 	TicketAddress []int
 }

 type Ticket struct {
 	ID int
 	OutletId int
 	PlayerId int
 	BuyTime int64
 	BuyNumber int

 	State int
 }

 func main() {
 	err := shim.Start(new(SimpleChaincode))
 	if err != nil {
 		fmt.Printf("Error starting Simple chaincode: %s", err)
 	}
 }

 // interface functions
 func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
 	
 	issuerBytes, err := t.CreateIssuer(stub, args)
 	if err != nil {
 		fmt.Printf("Failed to create issuer\n")
 		return nil, errors.New("init error" + err.Error())
 	}

 	lotteryBytes, err := t.CreateLottery(stub, args)
 	if err != nil {
 		fmt.Printf("Failed to create issuer\n")
 		return nil, errors.New("init error" + err.Error())
 	}
	
	return lotteryBytes, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	if function == "createOutlet" {
		return t.CreateOutlet(stub, args)
	} else if function == "createPlayer" {
		return t.CreatePlayer(stub, args)
	} else if function == "buyTicket" {
		return t.BuyTicket(stub, args)
	} 

	return nil,nil
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	if function == "getIssuer"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, issuerBytes, err := getIssuerById(stub,args[0])
		if err != nil {
			fmt.Println("Error get issuer")
			return nil, err
		}
		return issuerBytes, nil
	} else if function == "getLottery"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, lotteryBytes, err := getLotteryById(stub,args[0])
		if err != nil {
			fmt.Println("Error get lottery")
			return nil, err
		}
		return lotteryBytes, nil
	} else if function == "getOutlet"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, outletBytes, err := getOutletById(stub,args[0])
		if err != nil {
			fmt.Println("Error get outlet")
			return nil, err
		}
		return outletBytes, nil
	} else if function == "getPlayer"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, playerBytes, err := getPlayerById(stub,args[0])
		if err != nil {
			fmt.Println("Error get player")
			return nil, err
		}
		return playerBytes, nil
	} else if function == "getTicket"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, ticketBytes, err := getTicketById(stub,args[0])
		if err != nil {
			fmt.Println("Error get ticket")
			return nil, err
		}
		return ticketBytes, nil
	}

	return nil,nil
}

 // inner routines
 func GetAddress() (string, string, string) {
 	var address, privateKey, publicKey string
 	b := make([]byte, 48)

 	if _, err := io.ReadFull(rand.Reader, b); err != nil {
 		return "", "", ""
 	}

 	h := md5.New()
 	h.Write([]byte(base64.URLEncoding.EncodeToString(b)))

 	address = hex.EncodeToString(h.Sum(nil))
 	// TODO
 	privateKey = address + "1"
 	publicKey = address + "2"

 	return address, privateKey, publicKey
 }

 // FIXME TODO
 func getLuckyNumber() (int) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(TICKET_PERIOD))
	if err != nil {
        panic(err)
    }
    n := nBig.Int64()
	fmt.Printf("lucky number = %d\n", n)
    return int(n)
 }

// TODO
 func CheckSignature(PublicKey string, Signature string) (bool) {
 	// FIXME
 	A := PublicKey[:len(PublicKey)-1]
 	B := Signature[:len(Signature)-1]

 	return A == B
 }
// FIXME
 func checkTicket(stub shim.ChaincodeStubInterface, lottery Lottery, luckyNumber int) (int) {
 	var count = 0
 	for _, ticketAddress := range lottery.TicketAddress {
 		var ticketId = strconv.Itoa(ticketAddress)
 		ticket, _, error := getTicketById(stub, ticketId)
		if error != nil {
			fmt.Printf("Failed to get ticket %d\n", ticketAddress)
			continue
		}

		if ticket.BuyNumber == luckyNumber {
			ticket.State = WON
			count = count + 1
		} else {
			ticket.State = MISSED
		}

		var err = writeTicket(stub, ticket)
		if err != nil {
			fmt.Printf("Failed to write ticket %d\n", ticketAddress)
			// FIXME
			continue
		}		
 	}

 	fmt.Printf("Round %d got %d lucky dogs")

 	return count
 }

  func (t *SimpleChaincode) CreateIssuer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil,errors.New("Incorrect number of arguments. Expecting 1")
	}

	var issuer Issuer
	var issuerBytes []byte

	issuer = Issuer { ID:ISSUER_ID, Name:args[0] }
	err := writeIssuer(stub, issuer)
	if err != nil {
		return nil, errors.New("Write error" + err.Error())
	}

	issuerBytes, err = json.Marshal(&issuer)
	if err != nil {
		return nil, errors.New("Error retrieving issuerBytes")
	}

	return issuerBytes, nil
}

func (t *SimpleChaincode) CreateLottery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

 	var lottery Lottery
 	var lotteryBytes []byte
 	var ticketAddress []int

 	lottery = Lottery { ID:LOTTERY_ID, IssuerId:ISSUER_ID, Round:1, TicketAddress:ticketAddress }

 	err := writeLottery(stub, lottery)
 	if err != nil {
 		return nil, errors.New("write error" + err.Error())
 	}					

 	lotteryBytes, err = json.Marshal(&lottery)
 	if err != nil {
 		return nil, errors.New("Error retrieving lotteryBytes")
 	}

 	return lotteryBytes, nil
 }

func (t *SimpleChaincode) CreateOutlet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	var outlet Outlet
	var outletBytes []byte

	outlet = Outlet { ID:OutletNo, Name:args[0], Count:0 }
	err := writeOutlet(stub, outlet)
	if err != nil {
		return nil, errors.New("Write error" + err.Error())
	}

	outletBytes, err = json.Marshal(&outlet)
	if err != nil {
		return nil, errors.New("Error retrieving outletBytes")
	}

	OutletNo = OutletNo + 1

	return outletBytes, nil
}

 func (t *SimpleChaincode) CreatePlayer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil,errors.New("Incorrect number of arguments. Expecting 1")
	}

	var player Player
	var playerBytes []byte
	var ticketAddress []int

	player = Player { ID: PlayerNo, Name:args[0], TicketAddress:ticketAddress }
	err := writePlayer(stub, player)
	if err != nil{
		return nil, errors.New("Write error" + err.Error())
	}

	playerBytes, err = json.Marshal(&player)
	if err != nil {
		return nil, errors.New("Error retrieving playerBytes")
	}

	PlayerNo = PlayerNo + 1

	return playerBytes, nil
}

func (t *SimpleChaincode) BuyTicket(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	outletId, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for outlet")
	}

	playerId, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("Expecting integer value for player")
	}

	// check if lottery is valid
	var lotteryId = strconv.Itoa(LOTTERY_ID)
	lottery, _, error := getLotteryById(stub, lotteryId)
	if error != nil {
		return nil, errors.New("Error get lottery")
	}

	outlet, _, error := getOutletById(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get outlet")
	}	

	player, _, err := getPlayerById(stub, args[1])
	if err != nil {
		return nil, errors.New("Error get player")
	}
	
	// TODO:
	buyNumber := len(lottery.TicketAddress) + 1
	var ticket = Ticket { ID: TicketNo, OutletId:outletId, PlayerId:playerId, 
		BuyTime:time.Now().Unix(), BuyNumber:buyNumber, State:VALID }

	err = writeTicket(stub, ticket)
	if err != nil {
		return nil, errors.New("Error write ticket")
	}

	player.TicketAddress = append(player.TicketAddress, ticket.ID)
	err = writePlayer(stub, player)
	if err != nil {
		return nil, errors.New("Error write player")
	}

	lottery.TicketAddress = append(lottery.TicketAddress, ticket.ID)
	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write lottery")
	}

	ticketBytes, err := json.Marshal(&ticket)
	
	if err!= nil {
		return nil, errors.New("Error retrieving ticketBytes")
	}

	// draw lottery if condition is fulfilled
	if len(lottery.TicketAddress) == TICKET_PERIOD {
		drawLottery(stub, lottery)
	}

	TicketNo = TicketNo + 1
	return ticketBytes, nil
}

// helper routines

func drawLottery(stub shim.ChaincodeStubInterface, lottery Lottery) ([]byte, error) {

	var luckyNumber = getLuckyNumber()
	var ticketAddress []int

	_ = checkTicket(stub, lottery, luckyNumber)
	// FIXME: start a new round?
	lottery.Round = lottery.Round + 1
	lottery.TicketAddress = ticketAddress

	err := writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	lotteryBytes, err := json.Marshal(&lottery)
 	if err != nil {
 		return nil, errors.New("Error retrieving lotteryBytes")
 	}

 	return lotteryBytes, nil
}

func getIssuerById(stub shim.ChaincodeStubInterface, id string) (Issuer, []byte, error){
	var issuer Issuer
	issuerBytes, err := stub.GetState(ISSUER_PREFIX+id)
	if err != nil {
		fmt.Println("Error retrieving issuer")
		// TODO
	}

	err = json.Unmarshal(issuerBytes, &issuer)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return issuer, issuerBytes, nil
}

func getLotteryById(stub shim.ChaincodeStubInterface, address string) (Lottery, []byte, error) {
	var lottery Lottery
	lotteryBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(lotteryBytes, &lottery)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return lottery, lotteryBytes, nil
}

func getOutletById(stub shim.ChaincodeStubInterface, id string) (Outlet, []byte, error){
	var outlet Outlet
	outletBytes, err := stub.GetState(OUTLET_PREFIX+id)
	if err != nil {
		fmt.Println("Error retrieving outlet")
		// TODO
	}

	err = json.Unmarshal(outletBytes, &outlet)
	if err != nil {
		fmt.Println("Error unmarshalling outlet")
	}

	return outlet, outletBytes, nil
}

func getPlayerById(stub shim.ChaincodeStubInterface, id string) (Player, []byte, error){
	var player Player
	playerBytes, err := stub.GetState(PLAYER_PREFIX+id)
	if err != nil {
		fmt.Println("Error retrieving player")
		// TODO
	}

	err = json.Unmarshal(playerBytes, &player)
	if err != nil {
		fmt.Println("Error unmarshalling player")
	}

	return player, playerBytes, nil
}

func getTicketById(stub shim.ChaincodeStubInterface, id string) (Ticket, []byte, error){
	var ticket Ticket
	ticketBytes, err := stub.GetState(TICKET_PREFIX+id)
	if err != nil {
		fmt.Println("Error retrieving ticket")
		// TODO
	}

	err = json.Unmarshal(ticketBytes, &ticket)
	if err != nil {
		fmt.Println("Error unmarshalling ticket")
	}

	return ticket, ticketBytes, nil
}

func writeIssuer(stub shim.ChaincodeStubInterface, issuer Issuer) (error) {
	var issuerId = strconv.Itoa(issuer.ID)
	issuerBytes, err := json.Marshal(&issuer)
	if err != nil {
		return err
	}

	err = stub.PutState(ISSUER_PREFIX+issuerId, issuerBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeOutlet(stub shim.ChaincodeStubInterface, outlet Outlet) (error) {
	var outletId = strconv.Itoa(outlet.ID)
	outletBytes, err := json.Marshal(&outlet)
	if err != nil {
		return err
	}

	err = stub.PutState(OUTLET_PREFIX+outletId, outletBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeLottery(stub shim.ChaincodeStubInterface, lottery Lottery) (error) {
	var lotteryId = strconv.Itoa(lottery.ID)
	lotteryBytes, err := json.Marshal(&lottery)
	if err != nil{
	    return err
	}

	err = stub.PutState(LOTTERY_PREFIX+lotteryId, lotteryBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writePlayer(stub shim.ChaincodeStubInterface, player Player) (error) {
	var playerId = strconv.Itoa(player.ID)
	playerBytes, err := json.Marshal(&player)
	if err != nil {
		return err
	}

	err = stub.PutState(PLAYER_PREFIX+playerId, playerBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeTicket(stub shim.ChaincodeStubInterface, ticket Ticket) (error) {
	var ticketId = strconv.Itoa(ticket.ID)
	ticketBytes, err := json.Marshal(&ticket)
	if err != nil {
		return err
	}

	err = stub.PutState(TICKET_PREFIX+ticketId, ticketBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

