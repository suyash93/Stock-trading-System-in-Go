package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

type Args struct {
	StockpercentMap map[string]int
	Budget          float64
}

type PortfolioResponse struct {
	
	Stocksbought       string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

type Buyresponse struct {
	TradeID      int
	Stocksbought string
	UnvestedAmount float64
}

var X int

func main() {
	var stockinput string
	var Budget float64

	fmt.Printf("Enter stock symbol and Percentage: ")
	fmt.Scanln(&stockinput)
	fmt.Printf("Enter the budget ")
	fmt.Scanln(&Budget)

	sStocknum := strings.Split(stockinput, ",")
	count := 0
	StockpercentMap := make(map[string]int)
	for _, v := range sStocknum {
		sSplited := strings.Split(v, ":")
		sSplitnumper := strings.Split(sSplited[1], "%")
		i, err := strconv.Atoi(sSplitnumper[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		StockpercentMap[sSplited[0]] = i
		count = count + i
	}
	if count != 100 {
		fmt.Println("Total Percentage not 100")
		os.Exit(2)
	}
	args := &Args{StockpercentMap, Budget}

	var reply Buyresponse
	client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	c := jsonrpc.NewClient(client)
	err = c.Call("Stocking.BuyingStocks", args, &reply)
	if err != nil {
		log.Fatal("Error While Buying Stocks:", err)
	}
	fmt.Println("Trade id: ", reply.TradeID)
	fmt.Println("stocks: ", reply.Stocksbought)
	fmt.Println("unvested Amount: ", reply.UnvestedAmount)
    fmt.Println("")
	fmt.Printf("Enter the trading id ")
	fmt.Scanln(&X)
	var Port PortfolioResponse
	err = c.Call("Stocking.DisplayingPortfolio", &X, &Port)
	if err != nil {
		log.Fatal("Error While DisplayingPortfolio:", err)
	}
	fmt.Println("stocks: ", Port.Stocksbought)
	fmt.Println("current market value: ", Port.CurrentMarketValue)
	fmt.Println("unvested amount: ", Port.UnvestedAmount)

}
