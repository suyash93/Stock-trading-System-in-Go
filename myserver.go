package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
)


var tradingid = 0

var tradeMap = make(map[int]Trade)

type Args struct {
	Budget          float64
	StockpercentMap map[string]int
}

type Buyresponse struct {
	TradeID      int
	Stocksbought string
	UnvestedAmount float64
}

type PortfolioResponse struct {
	Stocksbought       string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

type Stock struct {
	name        string
	buyingPrice float64
	boughtcount int
}

type Trade struct {
	Tradingid int
	UnvestedAmount float64
	Stocks         []Stock
}

type Stock struct {
	List struct {
		Meta struct {
			Type  string `json:"type"`
			Start int    `json:"start"`
			Count int    `json:"count"`
		} `json:"meta"`
		Resources []struct {
			Resource struct {
				Classname string `json:"classname"`
				Fields    struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					Utctime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

type Stocking struct{}

func (t *Stocking) BuyingStocks(args *Args, reply *Buyresponse) error {
	buyCapacityMap := StockCalculation(args)
	priceMap := GetMarketPrice(buyCapacityMap)
	getstocks(priceMap, buyCapacityMap, reply)
	return nil
}

func (t *Stocking) DisplayingPortfolio(X *int,Port *PortfolioResponse) error {

	buyCapacityMap := make(map[string]float64)

	trade := tradeMap[(*X)]

	for istock := range trade.Stocks {
		buyCapacityMap[trade.Stocks[istock].name] = 0.00

	}
	priceMap := GetMarketPrice(buyCapacityMap)

	var buffer bytes.Buffer
	currMktPrice := 0.00
	for istock := range trade.Stocks {
		if istock > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(trade.Stocks[istock].name)
		buffer.WriteString(":")

		buffer.WriteString(strconv.Itoa(trade.Stocks[istock].boughtcount))
		buffer.WriteString(":")

		currPrice := priceMap[trade.Stocks[istock].name]
		currMktPrice = (currPrice * float64(trade.Stocks[istock].boughtcount)) + currMktPrice
		if currPrice > (trade.Stocks[istock].buyingPrice) {
			buffer.WriteString("+")
		} else if currPrice < (trade.Stocks[istock].buyingPrice) {
			buffer.WriteString("-")
		} else {
			buffer.WriteString("")
		}
		buffer.WriteString(strconv.FormatFloat(currPrice, 'f', 2, 64))

	}

	Port.CurrentMarketValue = currMktPrice
	Port.UnvestedAmount = trade.UnvestedAmount
	Port.Stocksbought = buffer.String()

	return nil
}

func StockCalculation(args *Args) map[string]float64 {

	buyCapacityMap := make(map[string]float64)

	for stock, percent := range args.StockpercentMap {
		buyCapacityMap[stock] = (float64(percent) / 100) * args.Budget
	}
	return buyCapacityMap
}

func getstocks(priceMap map[string]float64, buyCapacityMap map[string]float64, reply *Buyresponse) {
	unvested := 0.00
	var trade Trade
	var buffer bytes.Buffer
	counter := 0
	stockArr := make([]Stock, len(buyCapacityMap))
	for stock, capacity := range buyCapacityMap {
		
		price := priceMap[stock]
		

		
		if capacity > price {
			bought, _ := math.Modf(capacity / price)
			unvested = unvested + (capacity - (bought * price))
			stockArr[counter] = Stock{name: stock, buyingPrice: price, boughtcount: int(bought)}
			counter++
			if counter > 1 {
				buffer.WriteString(",")
			}
			buffer.WriteString(stock)
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(int(bought)))
			buffer.WriteString(":$")
			buffer.WriteString(strconv.FormatFloat(price, 'f', 2, 64))

		}


	}
	if counter == len(buyCapacityMap) {
		tradingid++
		trade.Tradingid = tradingid
		trade.UnvestedAmount = unvested
		trade.Stocks = stockArr
		tradeMap[tradingid] = trade
	}

	
	reply.TradeID = tradingid
	reply.UnvestedAmount = unvested
	reply.Stocksbought = buffer.String()
}

func GetMarketPrice(buyCapacityMap map[string]float64) map[string]float64 {
	var s Stock
	var priceMap map[string]float64
	var buffer bytes.Buffer
	buffer.WriteString("http://finance.yahoo.com/webservice/v1/symbols/")
	stockCounter := 0
	for stock := range buyCapacityMap {
		if stockCounter > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(stock)
		stockCounter++
	}

	buffer.WriteString("/quote?format=json")

	response, err := http.Get(buffer.String())
	if err != nil {
		fmt.Printf("error occured")
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		json.Unmarshal([]byte(contents), &s)
		priceMap = make(map[string]float64)
		for i := 0; i < s.List.Meta.Count; i++ {
			f, err1 := strconv.ParseFloat(s.List.Resources[i].Resource.Fields.Price, 64)
			priceMap[s.List.Resources[i].Resource.Fields.Symbol] = f
			if err1 != nil {
				fmt.Printf("%s", err1)
				os.Exit(1)
			}
		}
	}
	return priceMap
}

func main() {
	stk := new(Stocking)
	server := rpc.NewServer()
	server.Register(stk)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}