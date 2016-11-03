/**
 * @author: hsiung
 * @date: 2016/11/3
 * @desc: a simple lottery
 */

 package main
 import (
 	"errors"
 	"fmt"
 	"crypto/md5"
 	"crypto/rand"
 	"encoding/base64"
 	"encoding/hex"
 	"io"
 	"github.com/hyperledger/fabric/core/chaincode/shim"
 )

 type SimpleChainCode struct {

 }

 // global vars
 var targetBlock = 20000
 var ticketNo = 0

 type Lottery struct {
 	Name string
 	Address string
 	Type int
 	InitBalance int 	
 	EndTime int64
 	// ...
 	PrivateKey string
 	PublicKey string
 	TicketAddress []string
 }

 type Player struct {
 	Name string
 	Address string
 	PrivateKey string
 	PublicKey string
 }

 type Ticket struct {
 	Id int
 	Address string
 	LotteryAddress string
 	PlayerAddress string
 	PlayerSign string
 	BuyTime int64
 	BuyCount int
 	BuyCash	int
 	BuyNumber string
 }

 func main() {
 	err := shim.Start(new(SimpleChaincode))
 	if err != nil {
 		fmt.Printf("Error starting Simple chaincode: %s", err)
 	}
 }

 // inner routines
 func GetAddress() (string, string, string) {
 	var address, privateKey, publicKey string
 	b := make([]byte, 48)

 	if _, err := io.ReadFull(rand.Reader, b); err != nil {
 		return "", "", ""
 	}

 	h := md5:New()
 	h.Write([]byte(base64.URLEncoding.EncodeToString(b)))

 	address = hex.EncodeToString(h.Sum(nil))
 	// TODO
 	privateKey = address + "1"
 	publicKey = address + "2"

 	return address, privateKey, publicKey
 }

 func (t *SimpleChaincode) createLottery(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
 	if len(args) != 4 {
 		return nil, errors.New("Incorrect number of arguments, expecting 4")
 	}

 	var lottery Lottery
 	var lotteryBytes []byte
 	var ticketAddress []string
 	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

 	lottery = Lottery {Name:args[0], Address:address, Type:args[1], InitBalance:args[2], EndTime:args[3],
 						PrivateKey:privateKey, PublicKey:publicKey, TicketAddress:ticketAddress}

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

 func (t *SimpleChaincode) createPlayer(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil,errors.New("Incorrect number of arguments. Expecting 1")
	}

	var player Player
	var playerBytes []byte
	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

	player = Player {Name:args[0], Address:address, PrivateKey:privateKey, PublicKey:publicKey,}
	err := writePlayer(stub, player)
	if err != nil{
		return nil,errors.New("Write error" + err.Error())
	}

	playerBytes, err = json.Marshal(&player)
	if err != nil {
		return nil, errors.New("Error retrieving playerBytes")
	}

	return playerBytes, nil
}

func (t *SimpleChaincode) buyTicket(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	lottery, lotteryBytes, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	player, playerBytes, err:= getPlayerByAddress(stub, args[2])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	playerSign := args[1]
	var ticket Ticket
	var address string
	// FIXME
	address, _, _ = GetAddress()
	ticket = Ticket{Id:ticketNo, Address:address, LotteryAddress:args[0], PlayerAddress:args[2], PlayerSign:playerSign, BuyTime:time.Now().Unix(), BuyCount:args[3], BuyCash:args[4], BuyNumber:args[5]}

	err = writeTicket(stub, ticket)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	lottery.TicketAddress = append(lottery.TicketAddress, ticket.Address)
	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}
	// FIXME: update balance of player?
	// err = writePlayer(stub, player)
	// if err!= nil {
	// 	return nil,errors.New("Error write data")
	// }

	ticketNo = ticketNo + 1
	ticketBytes, err = json.Marshal(&ticket)
	
	if err!= nil {
		return nil,errors.New("Error retrieving ticketBytes")
	}

	return ticketBytes, nil
}

func getLotteryByAddress(stub *shim.ChaincodeStub, address string) (Lottery, []byte, error) {
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

func getTicketByAddress(stub *shim.ChaincodeStub, address string) (Ticket, []byte, error){
	var ticket Ticket
	ticketBytes, err := stub.GetState(address)
	if err != nil{
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(ticketBytes,&ticket)
	if err != nil{
		fmt.Println("Error unmarshalling data")
	}

	return ticket, ticketBytes, nil
}

func getTicketById(stub *shim.ChaincodeStub, id string) (Ticket, []]byte, error) {
	var ticket Ticket
	ticketBytes, err := stub.GetState("Ticket"+id)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(ticketBytes,&record)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return ticket, ticketBytes, nil
}

func writeTicket(stub *shim.ChaincodeStub, ticket Ticket) (error) {
	var ticketId string
	ticketBytes, err := json.Marshal(&ticket)
	if err != nil {
		return err
	}

	ticketId, _ = strconv.Itoa(ticket.Id)
	err = stub.PutState("Ticket"+ticketId, ticketBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeLottery(stub *shim.ChaincodeStub, lottery Lottery) (error) {
	lotteryBytes, err := json.Marshal(&lottery)
	if err != nil{
	    return err
	}

	err = stub.PutState(lottery.Address, lotteryBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writePlayer(stub *shim.ChaincodeStub, player Player) (error) {
	playerBytes, err := json.Marshal(&player)
	if err != nil {
		return err
	}

	err = stub.PutState(player.Address, playerBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

