package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gocolly/colly"
)

//Struct for scaping coinmarket data
type Coin struct {
	Rank              string
	Symbol            string
	Name              string
	Price             string
	PriceBTC          string
	Volume            string
	Marketcap         string
	CirculatingSupply string
	Change1hr         string
	Change24hr        string
	Change7d          string
}

//Struct to form final data once json structure has been formed from struct above
type CMCFormData []struct {
	CoinRank         string `json:"Rank"`
	Symbol           string `json:"Symbol"`
	Name             string `json:"Name"`
	PriceUSD         string `json:"Price"`
	PriceBTC         string `json:"PriceBTC"`
	VolumeUsd24h     string `json:"Volume"`
	MarketCapUsd     string `json:"Marketcap"`
	CircuSupply      string `json:"CirculatingSupply"`
	PercentChange1h  string `json:"Change1hr"`
	PercentChange24h string `json:"Change24hr"`
	PercentChange7d  string `json:"Change7d"`
}

const marketport = ":3099"

func integerToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func floatToString(value float64, precision int64) string {
	return strconv.FormatFloat(value, 'f', int(precision), 64)
}

func stringToFloat(value string) float64 {
	if value == "" || value == "None" || value == "?" {
		return 0
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func formatValue(key string, meta string, value string) string {
	result := key
	if meta != "" {
		result += "{" + meta + "}"
	}
	result += " "
	result += value
	result += "\n"
	return result
}

//function to scrap coin data from CoinMarketCap(CMC)
func scrapcmc() (string, error) {

	c := colly.NewCollector()

	c.AllowedDomains = []string{"coinmarketcap.com", "www.coinmarketcap.com"}

	coins := make([]Coin, 0)

	c.OnHTML("#currencies-all tbody tr", func(e *colly.HTMLElement) {

		tmpmarketcap := e.ChildAttr(".market-cap", "data-usd")

		if tmpmarketcap == "?" || tmpmarketcap == "None" || tmpmarketcap == "" {
			tmpmarketcap = "0"
		}

		finmarketcap := strconv.FormatFloat(stringToFloat(tmpmarketcap), 'f', -1, 64)

		coin := Coin{
			Rank:              e.ChildText("td.text-center"),
			Symbol:            e.ChildText(".col-symbol"),
			Name:              e.ChildText(".currency-name-container"),
			Price:             e.ChildAttr("a.price", "data-usd"),
			PriceBTC:          e.ChildAttr("a.price", "data-btc"),
			Volume:            e.ChildAttr("a.volume", "data-usd"),
			CirculatingSupply: e.ChildAttr(".circulating-supply > a", "data-supply"),
			Marketcap:         finmarketcap,
			Change1hr:         e.ChildAttr(".percent-1h", "data-usd"),
			Change24hr:        e.ChildAttr(".percent-24h", "data-usd"),
			Change7d:          e.ChildAttr(".percent-7d", "data-usd"),
		}

		coins = append(coins, coin)

	})

	c.Visit("https://coinmarketcap.com/all/views/all/")

	jsonData, err := json.MarshalIndent(coins, "", "  ")

	if err != nil {
		panic(err)
	}

	bodyString := string(jsonData)

	return bodyString, nil
}

// Function to set http server for prometheus metrics endpoint
func metrics(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /metrics")

	var up int64 = 1
	var jsonString string
	var err error

	jsonString, err = scrapcmc()
	if err != nil {
		log.Print(err)
		up = 0
	}

	jsonData := CMCFormData{}
	json.Unmarshal([]byte(jsonString), &jsonData)

	io.WriteString(w, formatValue("## Export coinmarketcap coins for prometheus indexing and analysis", "", integerToString(up)))

	for _, Coin := range jsonData {
		io.WriteString(w, formatValue("coin_rank", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.CoinRank), 0)))
		io.WriteString(w, formatValue("coin_price_usd", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.PriceUSD), 6)))
		io.WriteString(w, formatValue("coin_price_btc", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.PriceBTC), 6)))
		io.WriteString(w, formatValue("coin_24h_volume_usd", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.VolumeUsd24h), 1)))
		io.WriteString(w, formatValue("coin_market_cap_usd", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.MarketCapUsd), 3)))
		io.WriteString(w, formatValue("coin_circulating_supply", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.CircuSupply), 0)))
		io.WriteString(w, formatValue("coin_percent_change_1h", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.PercentChange1h), 2)))
		io.WriteString(w, formatValue("coin_percent_change_24h", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.PercentChange24h), 2)))
		io.WriteString(w, formatValue("coin_percent_change_7d", "symbol=\""+Coin.Symbol+"\",name=\""+Coin.Name+"\"", floatToString(stringToFloat(Coin.PercentChange7d), 2)))
	}
}

// Go webserver http index page to allow for navigation to prometheus metrics
func index(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /index")

	html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>CoinMarketCap Prometheus Exporter</title>
    </head>
    <body>
        <h1>CoinMarketCap Prometheus Exporter</h1>
        <p><a href="/metrics">Goto Metrics page</a></p>
    </body>
</html>`
	io.WriteString(w, html)
}

// Main function to server Site
func main() {

	log.Print("Scrapper Listening on port " + marketport)
	log.Print("Visit http://localhost" + marketport + "/metrics to view the metrics")
	log.Print("CTRL+C to cancel")
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(marketport, nil)

}
