package main

import (
	"HDWallet/database"
	"HDWallet/wallet"
	"fmt"
	"log"
)

func main() {
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	tokens := getToken()
	err = database.InsertTokenToTokenTable(db, tokens)
	if err != nil {
		log.Fatal(err)
	}
}

func getToken() []wallet.Token {
	numberAddress := 20

	//mnemonic, err := wallet.NewMnemonic()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(mnemonic)

	mnemonic := "veteran picture assist ginger gap poet denial crime art action palm damage"

	accoutnArray, err := wallet.GetWalletFromMnemonic(mnemonic, numberAddress)
	if err != nil {
		log.Fatal(err)
	}

	nonceSession, err := wallet.GetMultipleNonceAndSession(numberAddress)
	if err != nil {
		log.Fatal(err)
	}

	sig, err := wallet.GetMultipleSignature(nonceSession, accoutnArray)
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := wallet.GetMultipleToken(sig)
	if err != nil {
		log.Fatal(err)
	}

	for _, token := range tokens {
		tokenStr, err := token.String()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(tokenStr)
	}

	return tokens
}
