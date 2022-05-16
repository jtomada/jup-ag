package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/joho/godotenv"
	"github.com/jtomada/jup-ag"
)

var mintAddressMainnet = map[string]string{
	"SOL":  "So11111111111111111111111111111111111111112",
	"USDC": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
}

func main() {
	fmt.Println("Hello Jupiter!")

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	envWallet := os.Getenv("WALLET_PRIVATE_KEY")
	wallet, err := solana.PrivateKeyFromSolanaKeygenFile(envWallet)
	if err != nil {
		panic(err)
	}

	fmt.Println("wallet public key:", wallet.PublicKey().String())

	qr := jup.QuoteRequest{}
	qr.InputMint = mintAddressMainnet["SOL"]
	qr.OutputMint = mintAddressMainnet["USDC"]
	qr.Amount = 0.00001
	qr.Slippage = 1
	q, err := jup.GetQuote(&qr)
	if err != nil {
		panic(err)
	}

	sr := jup.SwapRequest{}
	sr.Route = q.Routes[0]
	sr.UserPublicKey = wallet.PublicKey().String()
	resp, err := jup.GetSwapTransactions(&sr)
	if err != nil {
		panic(err)
	}

	txs, err := decode(resp)
	if err != nil {
		panic(err)
	}

	err = sendTransactions(txs, wallet)
	if err != nil {
		panic(err)
	}
}

func decode(sr jup.SwapResponse) ([]solana.Transaction, error) {
	sertxs := [3]string{
		sr.SetupTransaction,
		sr.SwapTransaction,
		sr.CleanupTransaction,
	}

	resp := []solana.Transaction{}
	for _, sertx := range sertxs {
		if sertx != "" {
			tx, err := base64.StdEncoding.DecodeString(sertx)
			if err != nil {
				return nil, err
			}

			s := solana.MustTransactionFromDecoder(bin.NewBinDecoder(tx))
			resp = append(resp, *s)
		}
	}

	return resp, nil
}

func sendTransactions(txs []solana.Transaction, wallet solana.PrivateKey) error {
	rpcClient := rpc.New(rpc.MainNetBeta_RPC)
	wsClient, err := ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	if err != nil {
		return err
	}

	println("transaction count:", len(txs))

	for i, tx := range txs {
		recentBlockhash, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentConfirmed)
		if err != nil {
			return err
		}
		tx.Message.RecentBlockhash = recentBlockhash.Value.Blockhash

		// The serialized tx coming from Jupiter doesn't yet have a valid signature.
		tx.Signatures = []solana.Signature{}
		_, err = tx.Sign(
			func(key solana.PublicKey) *solana.PrivateKey {
				if wallet.PublicKey().Equals(key) {
					return &wallet
				}
				return nil
			},
		)
		if err != nil {
			return err
		}

		sig, err := confirm.SendAndConfirmTransactionWithOpts(
			context.TODO(),
			rpcClient,
			wsClient,
			&tx,
			false,
			rpc.CommitmentConfirmed,
		)
		if err != nil {
			return err
		}

		fmt.Println("tx signature:", i+1, sig.String())
	}
	return nil
}
