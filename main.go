package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Receiver struct {
	Address string
	Amount  string
}

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatalln("connection err: ", err.Error())
		return
	}

	api := ton.NewAPIClient(client)

	seedPhraseInput := "YOUR_ADDRESS_SEED_PHRASE"
	seedPhraseInput = strings.TrimSpace(seedPhraseInput)
	var seedPhrase *string
	if seedPhraseInput != "" {
		seedPhrase = &seedPhraseInput
	}

	// Initiate Wallet
	w := initiateWallet(seedPhrase, api)

	receiverAddress := "MINTER_CONTRACT_ADDRESS"
	receiverAddress = strings.TrimSpace(receiverAddress)
	var receiver *string
	if receiverAddress != "" {
		receiver = &receiverAddress
	}

	// Mint amount
	txAmount := 1000

	for i := 0; i < txAmount; i++ {
		log.Println("Sending transaction")
		if err := sendMessage(w, api, client, *receiver); err != nil {
			log.Println("Error sending messages:", err.Error())
		}
		log.Println("Sent", txAmount, "transactions")
		txAmount -= 1
	}

}

func initiateWallet(seedPhrase *string, api *ton.APIClient) *wallet.Wallet {
	var words []string

	if seedPhrase == nil {
		words = wallet.NewSeed()

	} else {
		words = strings.Split(*seedPhrase, " ")
	}

	w, err := wallet.FromSeed(api, words, wallet.V4R2)
	if err != nil {
		log.Fatalln("FromSeed err:", err.Error())
		return nil
	}

	log.Println("Wallet address:", w.Address())
	log.Println("Generated seed phrase:", strings.Join(words, " "))
	return w
}

func sendMessage(w *wallet.Wallet, api *ton.APIClient, client *liteclient.ConnectionPool, receiver string) error {
	ctx := client.StickyContext(context.Background())

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Println("CurrentMasterchainInfo err:", err.Error())
		return err
	}

	balance, err := w.GetBalance(context.Background(), block)
	if err != nil {
		log.Println("GetBalance err:", err.Error())
		return err
	}

	if balance.Nano().Uint64() >= 1.4e7 {
		log.Println("sending transaction and waiting for confirmation...")

		comm := "data:application/json,{\"p\":\"gram-20\",\"op\":\"mint\",\"tick\":\"gram\",\"repeat\":\"24\",\"amt\":\"10000\"}"
		body := cell.BeginCell().MustStoreInt(3501149081, 32).MustStoreStringSnake(comm).EndCell()

		transfer := &wallet.Message{
			Mode: 1,
			InternalMessage: &tlb.InternalMessage{
				Bounce:  false,
				DstAddr: address.MustParseAddr(receiver),
				Amount:  tlb.MustFromTON("0.008"),
				Body:    body,
			},
		}

		_, _, err = w.SendWaitTransaction(ctx, transfer)

		if err != nil {
			log.Println("Transfer err:", err.Error())
			return nil
		}

		time.Sleep(1 * time.Second)

		log.Printf("Transaction sent")
		return nil
	}
	log.Println("not enough balance:", balance.String())
	return nil
}
