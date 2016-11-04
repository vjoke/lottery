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
 var baseMoney = 5000000
 var targetBlock = 20000
 var ticketNo = 0
 var ticketPrice = 2
 var prizeUnit = 1000

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
 	// 
 	State int
 	TotalMoney int
 	InitiatorAddress string
 	TargetNumber string
 }

 type Authority struct {
 	Name string
 	Address string
 	PrivateKey string
 	PublicKey string
 	Money int
 }

 type Player struct {
 	Name string
 	Address string
 	PrivateKey string
 	PublicKey string
 	Money int
 }


 type Ticket struct {
 	Id int
 	Address string
 	LotteryAddress string
 	PlayerAddress string
 	PlayerSign string
 	BuyTime int64
 	Count int
 	BuyNumber string
 	// 0: INVALID, 1:VALID, 2:MISSED, 3:USED
 	State int
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
 	if len(args) != 5 {
 		return nil, errors.New("Incorrect number of arguments, expecting 5")
 	}

 	authority, authorityBytes, err := getAuthorityByAddress(stub, args[0])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	money, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Error parameter")
	}

	if money < baseMoney {
		return nil, erros.New("Not enough money")
	}

 	var lottery Lottery
 	var lotteryBytes []byte
 	var ticketAddress []string
 	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

 	lottery = Lottery {Name:args[1], Type:args[2], InitBalance:money, EndTime:args[4],
 						Address:address, PrivateKey:privateKey, PublicKey:publicKey, 
 						TicketAddress:ticketAddress, State:1, TotalMoney:money,
 						InitiatorAddress:authority.Address, TargetNumber:""}

 	// FIXME: update money
 	authority.Money = authority.Money - baseMoney
 	err := writeAuthority(stub, authority)
 	if err != nil {
 		return nil, erros.New("write error" + err.Error())
 	}

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

 func (t *SimpleChaincode) createAuthority(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}

	var authority Authority
	var authorityBytes []byte
	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

	authority = Authority { Name:args[0], Money:args[1], Address:address, PrivateKey:privateKey, PublicKey:publicKey }
	err := writeAuthority(stub, authority)
	if err != nil{
		return nil, errors.New("Write error" + err.Error())
	}

	authorityBytes, err = json.Marshal(&authority)
	if err != nil {
		return nil, errors.New("Error retrieving authorityBytes")
	}

	return authorityBytes, nil
}

 func (t *SimpleChaincode) createPlayer(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}

	var player Player
	var playerBytes []byte
	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

	player = Player { Name:args[0], Money:args[1], Address:address, PrivateKey:privateKey, PublicKey:publicKey }
	err := writePlayer(stub, player)
	if err != nil{
		return nil, errors.New("Write error" + err.Error())
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

	if lottery.State != 1 {
		return nil, errors.New("Lottery is not active")
	}

	player, playerBytes, err:= getPlayerByAddress(stub, args[2])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	playerSign := args[1]
	count, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Expecting integer value for count")
	}

	// update money
	var total int
	total = count * ticketPrice
	if player.Money < total {
		return nil, errors.New("Not enough money")
	}	

	var ticket Ticket
	var address string
	// FIXME
	address, _, _ = GetAddress()
	ticket = Ticket{Id:ticketNo, Address:address, 
		LotteryAddress:args[0], PlayerAddress:args[2], Count:count, BuyNumber:args[4],
		PlayerSign:playerSign, BuyTime:time.Now().Unix(), State:1}

	err = writeTicket(stub, ticket)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	lottery.TotalMoney = lottery.TotalMoney + total
	lottery.TicketAddress = append(lottery.TicketAddress, ticket.Address)
	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	player.Money = player.Money - total
	// FIXME: update balance of player
	err = writePlayer(stub, player)
	if err!= nil {
		return nil,errors.New("Error write data")
	}

	ticketNo = ticketNo + 1
	ticketBytes, err = json.Marshal(&ticket)
	
	if err!= nil {
		return nil,errors.New("Error retrieving ticketBytes")
	}

	return ticketBytes, nil
}

func (t *SimpleChaincode) takePrize(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	ticket, ticketBytes, error := getTicketByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	lottery, lotteryBytes, error := getLotteryByAddress(stub, ticket.LotteryAddress)
	if error != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State != 2 {
		return nil, errors.New("Lottery is active!")
	}

	player, playerBytes, err := getPlayerByAddress(stub, ticket.PlayerAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}
	// check signature
	playerSign := args[1]
	if playerSign != player.PlayerSign {
		return nil, errors.New("Not allowed")
	}

	if ticket.State == 3 {
		return nil, errors.New("Already taken the prize")
	}

	if ticket.BuyNumber != lottery.TargetNumber {
		return nil, errors.New("Missed")
	}

	// lucky man

	ticket.State = 3
	err = writeTicket(stub, ticket)
	if err != nil {
		return nil, errors.New("Error write data")
	}
	var totalPrize int
	totalPrize = lottery.count * prizeUnit
	// TODO: avoid underrun
	lottery.TotalMoney -= totalPrize

	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	player.Money = player.Money + totalPrize
	// FIXME: update balance of player
	err = writePlayer(stub, player)
	if err!= nil {
		return nil,errors.New("Error write data")
	}

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

func getAuthorityByAddress(stub *shim.ChaincodeStub, address string) (Authority, []byte, error){
	var authority Authority
	authorityBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(authorityBytes, &authority)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return authority, authorityBytes, nil
}

func getPlayerByAddress(stub *shim.ChaincodeStub, address string) (Player, []byte, error){
	var player Player
	playerBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(playerBytes,&player)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return player, playerBytes, nil
}

func getTicketByAddress(stub *shim.ChaincodeStub, address string) (Ticket, []byte, error){
	var ticket Ticket
	ticketBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(ticketBytes,&ticket)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return ticket, ticketBytes, nil
}

// func getTicketById(stub *shim.ChaincodeStub, id string) (Ticket, []]byte, error) {
// 	var ticket Ticket
// 	ticketBytes, err := stub.GetState("Ticket"+id)
// 	if err != nil {
// 		fmt.Println("Error retrieving data")
// 		// TODO
// 	}

// 	err = json.Unmarshal(ticketBytes, &record)
// 	if err != nil {
// 		fmt.Println("Error unmarshalling data")
// 	}

// 	return ticket, ticketBytes, nil
// }

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

func writeAuthority(stub *shim.ChaincodeStub, authority Authority) (error) {
	authorityBytes, err := json.Marshal(&authority)
	if err != nil {
		return err
	}

	err = stub.PutState(authority.Address, authorityBytes)
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

