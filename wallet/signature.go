package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
	"time"
)

type Message struct {
	Domain    string `json:"domain"`
	Address   string `json:"address"`
	Uri       string `json:"uri"`
	IssuedAt  string `json:"issued_at"`
	Statement string `json:"statement"`
	Nonce     string `json:"nonce"`
	ChainID   int    `json:"chain_id"`
}

type Signature struct {
	Wallet    Wallet `json:"wallet"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Time      string `json:"time"`
	Session   string `json:"session"`
}

func (msg *Message) String() (string, error) {
	message, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(message), nil
}

func (sig *Signature) String() (string, error) {
	signature, err := json.Marshal(sig)
	if err != nil {
		return "", err
	}

	return string(signature), nil
}

func GetMultipleSignature(nonceSession []NonceSession, wallets []Wallet) ([]Signature, error) {
	if nonceSession == nil || wallets == nil {
		return nil, fmt.Errorf("error: parameter must not be nil")
	}

	if len(nonceSession) != len(wallets) {
		return nil, fmt.Errorf("error: length of nonce array must be equal to length of wallet")
	}

	var sigArray []Signature

	for index, wallet := range wallets {
		now := time.Now().UTC()
		sig, err := GetSignatureByNonce(nonceSession[index].Nonce, wallet, now)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sigArray = append(sigArray, Signature{
			Wallet:    wallet,
			Signature: sig,
			Time:      now.Format(time.RFC3339),
			Nonce:     nonceSession[index].Nonce,
			Session:   nonceSession[index].Session,
		})
	}
	return sigArray, nil
}

func GetSignatureByNonce(nonce string, wallet Wallet, issuedAt time.Time) (string, error) {
	if issuedAt.IsZero() {
		issuedAt = time.Now().UTC()
	}

	message := Message{
		Domain:    "seal.gamefi.org",
		Address:   wallet.Address,
		Uri:       "https://seal.gamefi.org",
		IssuedAt:  issuedAt.Format(time.RFC3339),
		Statement: "Sign in with Ethereum",
		Nonce:     nonce,
		ChainID:   56,
	}

	sign, err := GetSignatureByMsg(message, wallet.PrivateKey)
	if err != nil {
		return "", nil
	}

	return sign, nil
}

func GetSignatureByMsg(msg Message, privateKey string) (string, error) {
	hashData, err := GetHashData(msg)
	if err != nil {
		return "", err
	}

	prvKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign(hashData, prvKey)
	if err != nil {
		return "", err
	}

	signature[64] += 27
	return hexutil.Encode(signature), nil
}

func GetHashData(msg Message) ([]byte, error) {
	message, err := siwe.InitMessage(msg.Domain, msg.Address, msg.Uri, msg.Nonce, map[string]interface{}{
		"issuedAt":  msg.IssuedAt,
		"statement": msg.Statement,
		"chainId":   msg.ChainID,
	})
	if err != nil {
		return nil, err
	}

	msgString := message.String()
	dataStringToHash := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msgString), msgString)
	hash := crypto.Keccak256Hash([]byte(dataStringToHash))
	return hash.Bytes(), nil
}
