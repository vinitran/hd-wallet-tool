package wallet

import (
	"encoding/json"
	"fmt"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type Wallet struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

func (w *Wallet) String() (string, error) {
	wallet, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	return string(wallet), nil
}

func NewMnemonic() (string, error) {
	entropy, err := hdwallet.NewEntropy(128)
	if err != nil {
		return "", err
	}

	mnemonic, err := hdwallet.NewMnemonicFromEntropy(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func GetWalletFromMnemonic(mnemonic string, numberAccount int) ([]Wallet, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	var wallets []Wallet
	for i := 0; i < numberAccount; i++ {
		derivationPath := fmt.Sprintf("m/44'/60'/0'/0/%d", i)
		wl, err := GetWalletFromDerivationPath(derivationPath, wallet)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wallets = append(wallets, *wl)
	}
	return wallets, nil
}

func GetWalletFromDerivationPath(derivationPath string, wallet *hdwallet.Wallet) (*Wallet, error) {
	path := hdwallet.MustParseDerivationPath(derivationPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		return nil, err
	}

	publicKey, err := wallet.PublicKeyHex(account)
	if err != nil {
		return nil, err
	}

	address, err := wallet.AddressHex(account)
	if err != nil {
		return nil, err
	}

	wl := Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}
	return &wl, nil
}
