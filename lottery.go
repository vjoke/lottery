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

 // global definitions
 const (
 	BASE_MONEY = 5000000
	PULISHMENT_MONEY = 100000
	TICKET_PRICE = 2
	MIN_COUNT = 1
	MAX_COUNT = 10
	DAYS_TO_CLOSE = 15
	FEE_PERCENT = 0.0000001
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
 	IssuerAddress string
 	LuckyNumber string
 	PrizeUnit int
 	CompanyAddress string
 }

 type Company struct {
 	Name string
 	Address string
 	PrivateKey string
 	PublicKey string
 	Money int
 }

 type Issuer struct {
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
 	OpenTime int64
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

 // interface functions
 func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
 	// function, args := stub.GetFunctionAndParameters()
	if function == "createCompany" {
		return t.createCompany(stub, args)
	} else if function == "createIssuer" {
		return t.createIssuer(stub, args)
	} else if function == "createPlayer" {
		return t.createPlayer(stub, args)
	}
	
	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// function, args := stub.GetFunctionAndParameters()
	if function == "createLottery" {
		return t.createLottery(stub, args)
	} else if function == "drawLottery" {
		return t.drawLottery(stub, args)
	} else if function == "closeLottery" {
		return t.closeLottery(stub, args)
	} else if function == "buyTicket" {
		return t.buyTicket(stub, args)
	} else if function == "takePrize" {
		return t.takePrize(stub, args)
	} 

	return nil,nil
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// function, args := stub.GetFunctionAndParameters()
	if function == "getCompanyByAddress"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, companyBytes, err := getCompanyByAddress(stub,args[0])
		if err != nil {
			fmt.Println("Error get company")
			return nil, err
		}
		return companyBytes, nil
	} else if function == "getIssuerByAddress"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, issuerBytes, err := getIssuerByAddress(stub,args[0])
		if err != nil {
			fmt.Println("Error get issuer")
			return nil, err
		}
		return issuerBytes, nil
	} else if function == "getPlayerByAddress"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, playerBytes, err := getPlayerByAddress(stub,args[0])
		if err != nil {
			fmt.Println("Error get player")
			return nil, err
		}
		return playerBytes, nil
	} else if function == "getLotteryByAddress"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, lotteryBytes, err := getLotteryByAddress(stub,args[0])
		if err != nil {
			fmt.Println("Error get lottery")
			return nil, err
		}
		return lotteryBytes, nil
	} else if function == "getTicketByAddress"{
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting 1")
		}
		_, ticketBytes, err := getTicketByAddress(stub,args[0])
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
 func GetLuckyNumber() (string) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
        panic(err)
    }
    n := nBig.Int64()
	fmt.Printf("lucky number = %d\n", n)
    return fmt.Sprintf("%d", n)
 }

// TODO
 func CheckSignature(PublicKey string, Signature string) (bool) {
 	// FIXME
 	A := PublicKey[:len(PublicKey)-1]
 	B := Signature[:len(Signature)-1]

 	return A == B
 }
// FIXME
 func CalcPrize(stub shim.ChaincodeStubInterface, lottery Lottery) (int) {
 	var prize = 0
 	var count = 0

 	for _, ticketAddress := range lottery.TicketAddress {
 		ticket, _, error := getTicketByAddress(stub, ticketAddress)
		if error != nil {
			continue
		}

		if ticket.BuyNumber == lottery.LuckyNumber {
			count = count + ticket.Count
		}
 	}

 	if count != 0 {
 		prize = lottery.InitMoney / count
 	}

 	return prize
 }

 func (t *SimpleChaincode) createLottery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

 	if len(args) != 6 {
 		return nil, errors.New("Incorrect number of arguments, expecting 6")
 	}

 	company, _, err := getCompanyByAddress(stub, args[4])
	if err != nil {
		return nil, errors.New("Error get data")
	}

 	issuer, _, err := getIssuerByAddress(stub, args[0])
	if err != nil {
		return nil, errors.New("Error get data")
	}

	issuerSign := args[5]
	if ! CheckSignature(issuer.PublicKey, issuerSign) {
		return nil, errors.New("Not allowed")
	}

	if issuer.Money < BASE_MONEY {
		return nil, errors.New("Not enough money")
	}

	endTime, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("Error parameter")
	}

	var lotteryType int
	var closeTime int64
 	var lottery Lottery
 	var lotteryBytes []byte
 	var ticketAddress []string
 	var address, privateKey, publicKey string

 	// TODO

 	lotteryType, _ = strconv.Atoi(args[2])

 	address, privateKey, publicKey = GetAddress()
 	closeTime = int64(endTime + DAYS_TO_CLOSE * 86400)

 	lottery = Lottery { Name:args[1], Type:lotteryType, InitMoney:BASE_MONEY, EndTime:int64(endTime), CloseTime:closeTime,
 						Address:address, PrivateKey:privateKey, PublicKey:publicKey, 
 						TicketAddress:ticketAddress, State:ACTIVE, TotalMoney:BASE_MONEY,
 						IssuerAddress:issuer.Address, LuckyNumber:"",
 						CompanyAddress:company.Address, PrizeUnit:0 }

 	// FIXME: update money
 	issuer.Money = issuer.Money - BASE_MONEY
 	err = writeIssuer(stub, issuer)
 	if err != nil {
 		return nil, errors.New("write error" + err.Error())
 	}

 	err = writeLottery(stub, lottery)
 	if err != nil {
 		return nil, errors.New("write error" + err.Error())
 	}					

 	lotteryBytes, err = json.Marshal(&lottery)
 	if err != nil {
 		return nil, errors.New("Error retrieving lotteryBytes")
 	}

 	return lotteryBytes, nil
 }

 func (t *SimpleChaincode) createCompany(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}

	var company Company
	var companyBytes []byte
	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

	company = Company { Name:args[0], Money:0, Address:address, PrivateKey:privateKey, PublicKey:publicKey }
	err := writeCompany(stub, company)
	if err != nil {
		return nil, errors.New("Write error" + err.Error())
	}

	companyBytes, err = json.Marshal(&company)
	if err != nil {
		return nil, errors.New("Error retrieving companyBytes")
	}

	return companyBytes, nil
}

 func (t *SimpleChaincode) createIssuer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}

	var issuer Issuer
	var issuerBytes []byte
	var address, privateKey, publicKey string

 	address, privateKey, publicKey = GetAddress()

 	money, _ := strconv.Atoi(args[1])
	issuer = Issuer { Name:args[0], Money:money, Address:address, PrivateKey:privateKey, PublicKey:publicKey }
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

 func (t *SimpleChaincode) createPlayer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil,errors.New("Incorrect number of arguments. Expecting 2")
	}

	var player Player
	var playerBytes []byte
	var address, privateKey, publicKey string

	money, _ := strconv.Atoi(args[1])
 	address, privateKey, publicKey = GetAddress()

	player = Player { Name:args[0], Money:money, Address:address, PrivateKey:privateKey, PublicKey:publicKey }
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

func (t *SimpleChaincode) buyTicket(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// check if lottery is valid
	lottery, _, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.EndTime > time.Now().Unix() {
		return nil, errors.New("Lottery is closed")
	}

	if lottery.State != ACTIVE {
		return nil, errors.New("Lottery is not active")
	}

	player, _, err := getPlayerByAddress(stub, args[2])
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
		PlayerSign:playerSign, BuyTime:time.Now().Unix(), OpenTime:0, State:VALID}

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

	ticketBytes, err := json.Marshal(&ticket)
	
	if err!= nil {
		return nil, errors.New("Error retrieving ticketBytes")
	}

	return ticketBytes, nil
}

func (t *SimpleChaincode) drawLottery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var timeNow int64
	timeNow = time.Now().Unix()

	lottery, lotteryBytes, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	issuer, _, err := getIssuerByAddress(stub, lottery.IssuerAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	issuerSign := args[1]
	if ! CheckSignature(issuer.PublicKey, issuerSign) {
		return nil, errors.New("Not allowed")
	}

	company, _, err := getCompanyByAddress(stub, lottery.CompanyAddress)
	if err != nil {
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
	lottery.PrizeUnit = CalcPrize(stub, lottery)
	lottery.State = DRAW
	// FIXME: to avoid cheating
	if lottery.EndTime < (timeNow - 86400) {
		lottery.TotalMoney = lottery.TotalMoney - PULISHMENT_MONEY
		company.Money = company.Money + PULISHMENT_MONEY
	}

	err = writeCompany(stub, company)
	if err != nil {
		return nil, errors.New("Error write data")
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

func (t *SimpleChaincode) takePrize(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var totalPrize int
	var timeNow int64
	timeNow = time.Now().Unix()

	ticket, ticketBytes, error := getTicketByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	lottery, _, err := getLotteryByAddress(stub, ticket.LotteryAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State == ACTIVE {
		return nil, errors.New("Lottery is active!")
	}

	if lottery.State == CLOSED {
		return nil, errors.New("Lottery is closed, too late!")
	}

	player, _, err := getPlayerByAddress(stub, ticket.PlayerAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}
	// TODO: check signature
	playerSign := args[1]
	if ! CheckSignature(player.PublicKey, playerSign) {
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
	ticket.OpenTime = timeNow
	totalPrize = ticket.Count * lottery.PrizeUnit
	// FIXME: avoid underrun
	lottery.TotalMoney = lottery.TotalMoney - totalPrize

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

func (t *SimpleChaincode) closeLottery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	lottery, _, error := getLotteryByAddress(stub, args[0])
	if error != nil {
		return nil, errors.New("Error get data")
	}

	company, _, err := getCompanyByAddress(stub, lottery.CompanyAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	if lottery.State != DRAW {
		return nil, errors.New("Lottery is active!")
	}

	issuer, issuerBytes, err := getIssuerByAddress(stub, lottery.IssuerAddress)
	if err != nil {
		return nil, errors.New("Error get data")
	}

	// TODO: check signature
	issuerSign := args[1]
	if ! CheckSignature(issuer.PublicKey, issuerSign) {
		return nil, errors.New("Not allowed")
	}

	if lottery.CloseTime < time.Now().Unix() {
		return nil, errors.New("Not expired")
	}
	// FIXME: close the lottery
	lottery.State = CLOSED
	var profit, fee int
	profit = lottery.TotalMoney - lottery.InitMoney
	if profit > 0 {
		var temp = float64(profit) * float64(FEE_PERCENT)
		fee = int(temp)
		profit = profit - fee
		company.Money = company.Money + fee
	} else {
		fee = 0
	}

	issuer.Money = issuer.Money + lottery.TotalMoney - fee
	// FIXME
	// lottery.TotalMoney = 0
	err = writeIssuer(stub, issuer)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	err = writeLottery(stub, lottery)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	err = writeCompany(stub, company)
	if err != nil {
		return nil, errors.New("Error write data")
	}

	issuerBytes, err = json.Marshal(&issuer)
	
	if err!= nil {
		return nil,errors.New("Error retrieving issuerBytes")
	}

	return issuerBytes, nil
}

func getLotteryByAddress(stub shim.ChaincodeStubInterface, address string) (Lottery, []byte, error) {
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

func getCompanyByAddress(stub shim.ChaincodeStubInterface, address string) (Company, []byte, error){
	var company Company
	companyBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(companyBytes, &company)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return company, companyBytes, nil
}

func getIssuerByAddress(stub shim.ChaincodeStubInterface, address string) (Issuer, []byte, error){
	var issuer Issuer
	issuerBytes, err := stub.GetState(address)
	if err != nil {
		fmt.Println("Error retrieving data")
		// TODO
	}

	err = json.Unmarshal(issuerBytes, &issuer)
	if err != nil {
		fmt.Println("Error unmarshalling data")
	}

	return issuer, issuerBytes, nil
}

func getPlayerByAddress(stub shim.ChaincodeStubInterface, address string) (Player, []byte, error){
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

func getTicketByAddress(stub shim.ChaincodeStubInterface, address string) (Ticket, []byte, error){
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

func writeTicket(stub shim.ChaincodeStubInterface, ticket Ticket) (error) {
	ticketBytes, err := json.Marshal(&ticket)
	if err != nil {
		return err
	}

	err = stub.PutState(ticket.Address, ticketBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeLottery(stub shim.ChaincodeStubInterface, lottery Lottery) (error) {
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

func writeCompany(stub shim.ChaincodeStubInterface, company Company) (error) {
	companyBytes, err := json.Marshal(&company)
	if err != nil {
		return err
	}

	err = stub.PutState(company.Address, companyBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writeIssuer(stub shim.ChaincodeStubInterface, issuer Issuer) (error) {
	issuerBytes, err := json.Marshal(&issuer)
	if err != nil {
		return err
	}

	err = stub.PutState(issuer.Address, issuerBytes)
	if err != nil {
		return errors.New("PutState error" + err.Error())
	}

	return nil
}

func writePlayer(stub shim.ChaincodeStubInterface, player Player) (error) {
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

