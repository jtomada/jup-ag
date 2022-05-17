package jup

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

const swapUrl = "/v1/swap"
const quoteUrl = "/v1/quote"
const priceUrl = "/v1/price"
const mapUrl = "v1/indexed-route-map"
const baseUrl = "https://quote-api.jup.ag"

func GetSwapTransactions(swap *SwapRequest) (*SwapResponse, error) {
	// Get the serialized transaction(s) from Jupiter's Swap API
	var jsonBody bytes.Buffer
	err := json.NewEncoder(&jsonBody).Encode(&swap)
	if err != nil {
		return nil, err
	}

	url := baseUrl + swapUrl

	r, err := http.Post(url, "application/json", &jsonBody)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	s := &SwapResponse{}
	err = json.NewDecoder(r.Body).Decode(s)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", *s)

	return s, nil
}

type QuoteRequest struct {
	InputMint        string
	OutputMint       string
	Amount           float64
	Slippage         float64
	FeeBps           float64
	OnlyDirectRoutes bool
}

func GetQuote(qr *QuoteRequest) (*Quote, error) {
	qurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	qurl.Path += quoteUrl

	amountLamports := qr.Amount * 1000000000
	a := fmt.Sprintf("%f", amountLamports)
	s := fmt.Sprintf("%f", qr.Slippage)
	f := fmt.Sprintf("%f", qr.FeeBps)

	params := url.Values{}
	params.Add("inputMint", qr.InputMint)
	params.Add("outputMint", qr.OutputMint)
	params.Add("amount", a)
	params.Add("slippage", s)
	params.Add("feeBps", f)

	if qr.OnlyDirectRoutes {
		params.Add("onlyDirectRoutes", "true")
	} else {
		params.Add("onlyDirectRoutes", "false")
	}

	qurl.RawQuery = params.Encode()
	fmt.Printf("Encoded URL is %q\n", qurl.String())

	r, err := http.Get(qurl.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	quote := &Quote{}
	err = json.NewDecoder(r.Body).Decode(quote)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", *quote)

	return quote, nil
}

type IndexedRouteMapResponse struct {
	MintKeys        []string         `json:"mintKeys"`
	IndexedRouteMap map[string][]int `json:"indexedRouteMap"`
}

func GetIndexedRouteMap(onlyDirectRoutes bool) (*IndexedRouteMapResponse, error) {
	murl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	murl.Path += mapUrl

	params := url.Values{}
	if onlyDirectRoutes {
		params.Add("onlyDirectRoutes", "true")
	} else {
		params.Add("onlyDirectRoutes", "false")
	}

	r, err := http.Get(murl.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	imap := &IndexedRouteMapResponse{}
	err = json.NewDecoder(r.Body).Decode(imap)
	if err != nil {
		return nil, err
	}
	fmt.Printf("len route map: %d\n", len(imap.IndexedRouteMap))

	return imap, nil
}

type Price struct {
	Data      PriceData `json:"data"`
	TimeTaken float64   `json:"timeTaken"`
}

type PriceData struct {
	InputMint    string  `json:"inputMint"`
	InputSymbol  string  `json:"inputSymbol"`
	OutputMint   string  `json:"outputMint"`
	OutputSymbol string  `json:"outputSymbol"`
	Amount       int     `json:"amount"`
	Price        float64 `json:"price"`
}

type PriceRequest struct {
	InputMint  string
	OutputMint string
	Amount     float64
}

func GetPrice(p *PriceRequest) (*Price, error) {
	purl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	purl.Path += priceUrl

	amountLamports := p.Amount * 1000000000
	a := fmt.Sprintf("%f", amountLamports)

	params := url.Values{}
	params.Add("inputMint", p.InputMint)
	params.Add("outputMint", p.OutputMint)
	params.Add("amount", a)

	purl.RawQuery = params.Encode()
	fmt.Printf("Encoded URL is %q\n", purl.String())

	r, err := http.Get(purl.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	price := &Price{}
	err = json.NewDecoder(r.Body).Decode(price)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", *price)

	return price, nil
}
