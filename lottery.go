/**
 * @author: hsiung
 * @date: 2016/11/3
 * @desc: a simple lottery
 */

 package main
 import (
 	"errors"
 	"fmt"
 	"math/rand"
 	"crypto/md5"
 	"crypto/rand"
 	"encoding/base64"
 	"encoding/hex"
 	"io"
 	"github.com/hyperledger/fabric/core/chaincode/shim"
 )

 type SimpleChainCode struct {

 }

 // global definitions
 const (
 	BASE_MONEY = 5000000
	PULISHMENT_MONEY = 100000
	TICKET_PRICE = 2
	PRIZE_UNIT = 1000
	MIN_COUNT = 1
	MAX_COUNT = 10
	DAYS_TO_CLOSE = 15
 )

// lottery state
 const (
 	ACTIVE 	= 1
 	DRAW 	= 2
 	CLOSED 	= 3
 )

 // ticket state
 const (
 	INVALID = 0
 	VALID 	= 1
 	MISSED 	= 2
 	WON 	= 3
 )

 type Lottery struct {
 	Name string
 	Address string
 	Type int
 	InitMoney int 	
 	EndTime int64
 	CloseTime int64
 	
 	PrivateKey string
 	PublicKey string
 	TicketAddress []string
 	
 	State int
 	TotalMoney int
 	AuthorityAddress string
 	LuckyNumber string
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
 	Address string
 	LotteryAddress string
 	PlayerAddress string
 	PlayerSign string
 	BuyTime int64
 	Count int
 	BuyNumber string
 	
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

 // FIXME
 func GetLuckyNumber() (string) {
 	rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    n := r1.Int(1000000)

    return fmt.Sprintf("%d", n)
 }

// TODO
 func CheckSignature(PublicKey string, Signature string) (bool) {
 	// FIXME
 	A := PublicKey[:len(PublicKey)-1]
 	B := Signature[:len(Signature)-1]

 	return A == B
 }

 func (t *SimpleChaincode) createLottery(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

 	if len(args) != 5 {
 		return nil, errors.New("Incorrect number of arguments, expecting 5")
 	}

 	authority, authorityBytes, err := getAuthorityByAddress(stub, args[0])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	money, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Error parameter")
	}

	if money < BASE_MONEY {
		return nil, erros.New("Not enough money")
	}

	endTime, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New("Error parameter")
	}

	var closeTime int64
 	var lottery Lottery
 	var lotteryBytes []byte
 	var ticketAddress []string
 	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()
 	closeTime = endTime + DAYS_TO_CLOSE * 86400

 	lottery = Lottery { Name:args[1], Type:args[2], InitMoney:money, EndTime:endTime, CloseTime:closeTime
 						Address:address, PrivateKey:privateKey, PublicKey:publicKey, 
 						TicketAddress:ticketAddress, State:ACTIVE, TotalMoney:money,
 						AuthorityAddress:authority.Address, LuckyNumber:"" }

 	// FIXME: update money
 	authority.Money = authority.Money - BASE_MONEY
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
	if err != nil {
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
	// check if lottery is valid
	lottery, lotteryBytes, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.EndTime > time.Now().Unix() {
		return nil, errors.New("Lottery is closed")
	}

	if lottery.State != ACTIVE {
		return nil, errors.New("Lottery is not active")
	}

	player, playerBytes, err := getPlayerByAddress(stub, args[2])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	playerSign := args[1]
	count, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Expecting integer value for count")
	}

	var total int
	total = count * TICKET_PRICE
	if player.Money < total {
		return nil, errors.New("Not enough money")
	}

	// TODO: check range
	var ticket Ticket
	var address string
	// FIXME
	address, _, _ = GetAddress()
	ticket = Ticket { Address:address, 
		LotteryAddress:args[0], PlayerAddress:args[2], Count:count, BuyNumber:args[4],
		PlayerSign:playerSign, BuyTime:time.Now().Unix(), State:VALID}

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

	ticketBytes, err = json.Marshal(&ticket)
	
	if err!= nil {
		return nil, errors.New("Error retrieving ticketBytes")
	}

	return ticketBytes, nil
}

func (t *SimpleChaincode) drawLottery(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var timeNow int64
	timeNow = time.Now().Unix()

	lottery, lotteryBytes, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State != ACTIVE {
		return nil, errors.New("Lottery is not active!")
	}
	// FIXME: time is not a good idea, use block height to trigger?
	if lottery.EndTime < timeNow {
		return nil, errors.New("Lottery is still active")
	}

	lottery.LuckyNumber = GetLuckyNumber()
	lottery.State = DRAW
	// FIXME: to avoid cheating
	if lottery.EndTime < (timeNow - 86400) {
		lottery.TotalMoney = lottery.TotalMoney - PULISHMENT_MONEY
	}

	lottery.CloseTime = timeNow + DAYS_TO_CLOSE*86400

	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	lotteryBytes, err = json.Marshal(&lottery)
 	if err != nil {
 		return nil, errors.New("Error retrieving lotteryBytes")
 	}

 	return lotteryBytes, nil
}

func (t *SimpleChaincode) takePrize(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var totalPrize int
	var timeNow int64
	timeNow = time.Now().Unix()

	ticket, ticketBytes, error := getTicketByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	lottery, lotteryBytes, err := getLotteryByAddress(stub, ticket.LotteryAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State == ACTIVE {
		return nil, errors.New("Lottery is active!")
	}

	if lottery.State == CLOSED {
		return nil, errors.New("Lottery is closed, too late!")
	}

	player, playerBytes, err := getPlayerByAddress(stub, ticket.PlayerAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}
	// TODO: check signature
	playerSign := args[1]
	if ! CheckSignature(player.publicKey, playerSign) {
		return nil, errors.New("Not allowed")
	}

	if ticket.State == WON {
		return nil, errors.New("Already taken the prize")
	}

	if ticket.BuyNumber != lottery.LuckyNumber {
		return nil, errors.New("Missed")
	}

	// TODO
	ticket.State = WON
	totalPrize = ticket.Count * PRIZE_UNIT
	// TODO: avoid underrun
	lottery.TotalMoney -= totalPrize

	err = writeTicket(stub, ticket)
	if err != nil {
		return nil, errors.New("Error write data")
	}

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

func (t *SimpleChaincode) closeLottery(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	lottery, lotteryBytes, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State != DRAW {
		return nil, errors.New("Lottery is active!")
	}

	authority, authorityBytes, err := getAuthorityByAddress(stub, lottery.AuthorityAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	// TODO: check signature
	authoritySign := args[1]
	if ! CheckSignature(authority.publicKey, authoritySign) {
		return nil, errors.New("Not allowed")
	}

	if lottery.CloseTime < time.Now().Unix() {
		return nil, errors.New("Not expired")
	}
	// close the lottery
	lottery.State = CLOSED
	authority.Money = authority.Money + lottery.TotalMoney
	// FIXME
	// lottery.TotalMoney = 0
	err = writeAuthority(stub, authority)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	authorityBytes, err = json.Marshal(&authority)
	
	if err!= nil {
		return nil,errors.New("Error retrieving authorityBytes")
	}

	return authorityBytes, nil
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

