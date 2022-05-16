package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SwapRequest struct {
	Route         Route  `json:"route"`
	WrapUnwrapSOL bool   `json:"wrapUnwrapSOL,omitempty"`
	FeeAccount    string `json:"feeAccount,omitempty"`
	TokenLedger   string `json:"tokenLedger,omitempty"`
	UserPublicKey string `json:"userPublicKey"`
}

type SwapResponse struct {
	SetupTransaction   string `json:"setupTransaction,omitempty"`
	SwapTransaction    string `json:"swapTransaction"`
	CleanupTransaction string `json:"cleanupTransaction,omitempty"`
}

type Quote struct {
	Routes    []Route `json:"data"`
	TimeTaken float64 `json:"timeTaken"`
}

type Route struct {
	InAmount              float64      `json:"inAmount"`
	OutAmount             float64      `json:"outAmount"`
	OutAmountWithSlippage float64      `json:"outAmountWithSlippage"`
	PriceImpactPct        float64      `json:"priceImpactPct"`
	MarketInfos           []MarketInfo `json:"marketInfos"`
}

type MarketInfo struct {
	ID                 string  `json:"id"`
	Label              string  `json:"label"`
	InputMint          string  `json:"inputMint"`
	OutputMint         string  `json:"outputMint"`
	NotEnoughLiquidity bool    `json:"notEnoughLiquidity"`
	InAmount           float64 `json:"inAmount"`
	OutAmount          float64 `json:"outAmount"`
	PriceImpactPct     float64 `json:"priceImpactPct"`
	LpFee              Fee     `json:"lpFee"`
	PlatformFee        Fee     `json:"platformFee"`
}

type Fee struct {
	Amount float64 `json:"amount"`
	Mint   string  `json:"mint"`
	Pct    float64 `json:"pct"`
}

var mintAddressMainnet = map[string]string{
	"SOL":  "So11111111111111111111111111111111111111112",
	"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
}

func main() {
	fmt.Println("Hello Jupiter!")

	// Get the best routes from Jupiter's Swap API
	quoteUrl, err := url.Parse("https://quote-api.jup.ag")
	if err != nil {
		panic(err)
	}

	quoteUrl.Path += "/v1/quote"

	params := url.Values{}
	params.Add("inputMint", mintAddressMainnet["SOL"])
	params.Add("outputMint", mintAddressMainnet["USDC"])
	params.Add("amount", "1000")
	params.Add("slippage", "0.5")
	quoteUrl.RawQuery = params.Encode()
	fmt.Printf("Encoded URL is %q\n", quoteUrl.String())

	resp, err := http.Get(quoteUrl.String())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	quote := Quote{}
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", quote)

	// Get the serialized transaction(s) from Jupiter's Swap API
	swapUrl := "https://quote-api.jup.ag/v1/swap"

	swapReq := SwapRequest{}
	swapReq.Route = quote.Routes[0]
	swapReq.UserPublicKey = wallet.PublicKey().String()

	var swapJsonBody bytes.Buffer
	err = json.NewEncoder(&swapJsonBody).Encode(&swapReq)
	if err != nil {
		panic(err)
	}

	resp, err = http.Post(swapUrl, "application/json", &swapJsonBody)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	swapResp := SwapResponse{}
	err = json.NewDecoder(resp.Body).Decode(&swapResp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", swapResp)
}
