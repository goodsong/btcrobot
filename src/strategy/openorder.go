/*
  btcrobot is a Bitcoin, Litecoin and Altcoin trading bot written in golang,
  it features multiple trading methods using technical analysis.

  Disclaimer:

  USE AT YOUR OWN RISK!

  The author of this project is NOT responsible for any damage or loss caused
  by this software. There can be bugs and the bot may not perform as expected
  or specified. Please consider testing it first with paper trading /
  backtesting on historical data. Also look at the code to see what how
  it's working.

  Weibo:http://weibo.com/bocaicfa
*/

package strategy

import (
	. "common"
	. "config"
	"fmt"
	"logger"
	"strconv"
	"time"
)

type OOStrategy struct {
	PrevKDJTrade string
	PrevTime     string
	PrevPrice    float64
	PrevBuyPirce float64
	BuyId        []string
	SellId       []string
	BuyBegin     time.Time
	SellBegin    time.Time
}

func init() {
	oo := new(OOStrategy)
	oo.PrevKDJTrade = "init"

	Register("OPENORDER", oo)
}

//KDJ-EX strategy
func (oo *OOStrategy) Perform(tradeAPI TradeAPI, records []Record) bool {

	const btcslap = 0.2
	const ltcslap = 0.01
	const timeout = 10
	const ordercount = 3

	numTradeAmount, err := strconv.ParseFloat(Option["tradeAmount"], 64)
	if err != nil {
		logger.Errorln("config item tradeAmount is not float")
		return false
	}

	nSplitTradeAmount := numTradeAmount / float64(ordercount)
	splitTradeAmount := fmt.Sprintf("%f", nSplitTradeAmount)

	var Time []string
	var Price []float64
	var Volumn []float64
	for _, v := range records {
		Time = append(Time, v.TimeStr)
		Price = append(Price, v.Close)
		Volumn = append(Volumn, v.Volumn)
		//Price = append(Price, (v.Close+v.Open+v.High+v.Low)/4.0)
		//Price = append(Price, v.Low)
	}

	ret, orderbook := tradeAPI.GetOrderBook()
	if !ret {
		logger.Infoln("get orderbook failed 1")
		ret, orderbook = tradeAPI.GetOrderBook() //try again
		if !ret {
			logger.Infoln("get orderbook failed 2")
			return false
		}
	}

	logger.Infoln("卖一", orderbook.Asks[len(orderbook.Asks)-1])

	logger.Infoln("买一", orderbook.Bids[0])

	var flag float64
	if orderbook.Bids[0].Price+0.02 > orderbook.Asks[len(orderbook.Asks)-1].Price {
		flag = 0
	} else {
		flag = 0.01
	}
	for i := 1; i <= ordercount; i++ {
		warning := "oo, 买入buy In<----限价单"
		tradePrice := fmt.Sprintf("%f", orderbook.Bids[0].Price+flag)
		buyID := tradeAPI.Buy(tradePrice, splitTradeAmount)
		if buyID != "0" {
			warning += "[委托成功]"
			oo.BuyId = append(oo.BuyId, buyID)
		} else {
			warning += "[委托失败]"
		}

		logger.Infoln(warning)

		warning := "oo, 卖出Sell Out---->限价单"
		tradePrice := fmt.Sprintf("%f", orderbook.Asks[len(orderbook.Asks)-1].Price-flag)
		sellID := tradeAPI.Sell(tradePrice, splitTradeAmount)
		if sellID != "0" {
			warning += "[委托成功]"
			oo.SellId = append(oo.SellId, sellID)
		} else {
			warning += "[委托失败]"
		}

		logger.Infoln(warning)
	}

	//check timeout trade
	now := time.Now()

	time.Sleep(10 * time.Second)
	logger.Infoln("time go ", int64(now.Sub(oo.BuyBegin)/time.Second))
	logger.Infoln("BuyId len", len(oo.BuyId), cap(oo.BuyId))
	logger.Infoln("SellId len", len(oo.SellId), cap(oo.SellId))

	if len(oo.BuyId) != 0 &&
		int64(now.Sub(oo.BuyBegin)/time.Second) > timeout {
		//todo-
		for _, BuyId := range oo.BuyId {
			warning := "<--------------buy order timeout, cancel-------------->" + BuyId
			if tradeAPI.CancelOrder(BuyId) {
				warning += "[Cancel委托成功]"
			} else {
				warning += "[Cancel委托失败]"
			}
			logger.Infoln(warning)
			time.Sleep(1 * time.Second)
		}
		oo.BuyId = oo.BuyId[:0]
	}

	if len(oo.SellId) != 0 &&
		int64(now.Sub(oo.SellBegin)/time.Second) > timeout {
		//todo
		for _, SellId := range oo.SellId {
			warning := "<--------------sell order timeout, cancel------------->" + SellId
			if tradeAPI.CancelOrder(SellId) {
				warning += "[Cancel委托成功]"
			} else {
				warning += "[Cancel委托失败]"
			}
			logger.Infoln(warning)
			time.Sleep(1 * time.Second)
		}
		oo.SellId = oo.SellId[:0]
	}

	return true
}
