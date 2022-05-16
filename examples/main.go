package main

import (
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
	"github.com/jtomada/jup-ag"
)

var mintAddressMainnet = map[string]string{
	"SOL":  "So11111111111111111111111111111111111111112",
	"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
}

func main() {
	fmt.Println("Hello Jupiter!")

	qr := jup.QuoteRequest{}
	qr.InputMint = mintAddressMainnet["SOL"]
	qr.OutputMint = mintAddressMainnet["USDC"]
	qr.Amount = 0.001
	qr.Slippage = 0.5
	q, err := jup.GetQuote(&qr)
	if err != nil {
		panic(err)
	}

	sr := jup.SwapRequest{}
	sr.Route = q.Routes[0]

	err = godotenv.Load()
	if err != nil {
		panic(err)
	}
	envWallet := os.Getenv("WALLET_PRIVATE_KEY")
	wallet, err := solana.PrivateKeyFromSolanaKeygenFile(envWallet)
	if err != nil {
		panic(err)
	}
	fmt.Println("wallet public key:", wallet.PublicKey().String())

	sr.UserPublicKey = wallet.PublicKey().String()
	_, err = jup.GetSwapTransactions(&sr)
	if err != nil {
		panic(err)
	}
}
