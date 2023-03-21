package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type NonceSession struct {
	Nonce   string `json:"nonce"`
	Session string `json:"session"`
}

type Token struct {
	Token     string    `json:"token"`
	Signature Signature `json:"signature"`
}

type MessageToPost struct {
	Domain    string `json:"domain"`
	Address   string `json:"address"`
	Uri       string `json:"uri"`
	Version   string `json:"version"`
	ChainId   int    `json:"chainId"`
	Nonce     string `json:"nonce"`
	Statement string `json:"statement"`
	IssuedAt  string `json:"issuedAt"`
}

func (n *NonceSession) String() (string, error) {
	data, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (n *Token) String() (string, error) {
	data, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func GetMultipleNonceAndSession(number int) ([]NonceSession, error) {
	var nonceSession []NonceSession

	// create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// create a channel to limit the number of goroutines
	maxGoroutines := 10
	goroutineChan := make(chan struct{}, maxGoroutines)
	for i := 0; i < number; i++ {
		// add a new goroutine to the wait group
		wg.Add(1)

		// add a marker to the goroutine channel to limit the number of running goroutines
		goroutineChan <- struct{}{}
		go func() {
			nonses, err := GetNonceAndSession()
			if err != nil {
				fmt.Println(err)
				<-goroutineChan
				wg.Done()
				return
			}
			nonceSession = append(nonceSession, *nonses)

			// remove the marker from the goroutine channel to free up space for another goroutine
			<-goroutineChan

			// signal to the wait group that this goroutine is finished
			wg.Done()
		}()
	}
	wg.Wait()
	return nonceSession, nil
}

func GetNonceAndSession() (*NonceSession, error) {
	response, err := http.Get("https://seal.gamefi.org/auth/siwe/nonce")

	if err != nil {
		return nil, err
	}
	cookie := strings.Split(response.Header.Get("set-cookie"), "; ")
	// Cookie[0] is a string that contain sub-string "session=". To get session, we must get from the eighth letter
	if len(cookie) < 1 || len(cookie[0]) < 8 {
		return nil, err
	}
	session := cookie[0][8:len(cookie[0])]

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Data string `json:"data"`
	}
	var responseObject Response

	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		return nil, err
	}

	nonce := responseObject.Data
	nonceSession := NonceSession{
		Nonce:   nonce,
		Session: session,
	}
	return &nonceSession, nil
}

func GetMultipleToken(sigs []Signature) ([]Token, error) {
	var tokenData []Token
	// create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// create a channel to limit the number of goroutines
	maxGoroutines := 10
	goroutineChan := make(chan struct{}, maxGoroutines)
	for _, sig := range sigs {
		// add a new goroutine to the wait group
		wg.Add(1)

		// add a marker to the goroutine channel to limit the number of running goroutines
		goroutineChan <- struct{}{}
		go func(sig Signature) {
			token, err := GetToken(sig)
			if err != nil {
				<-goroutineChan
				wg.Done()
				return
			}
			tokenData = append(tokenData, *token)
			<-goroutineChan
			wg.Done()
		}(sig)
	}

	wg.Wait()
	return tokenData, nil
}

func GetToken(sig Signature) (*Token, error) {
	msg := MessageToPost{
		Domain:    "seal.gamefi.org",
		Address:   sig.Wallet.Address,
		Uri:       "https://seal.gamefi.org",
		Version:   "1",
		ChainId:   56,
		Nonce:     sig.Nonce,
		Statement: "Sign in with Ethereum",
		IssuedAt:  sig.Time,
	}

	type TokenResponse struct {
		Data string `json:"data"`
	}

	token, err := postMessage(msg, sig.Signature, sig.Session)
	if err != nil {
		return nil, err
	}

	var responseToken TokenResponse
	err = json.Unmarshal(token, &responseToken)
	if err != nil {
		return nil, err
	}

	data := Token{
		Token:     responseToken.Data,
		Signature: sig,
	}

	return &data, nil
}

func postMessage(msg MessageToPost, sig, session string) ([]byte, error) {
	type PostData struct {
		Message   MessageToPost `json:"message"`
		Signature string        `json:"signature"`
	}

	message, err := json.Marshal(PostData{
		Message:   msg,
		Signature: sig,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", "https://seal.gamefi.org/auth/siwe/login", bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}

	r.Header.Add("x-captcha-internal", "wU7gcfS5aj2jxw9p")
	cookie := fmt.Sprintf("session=%s", session)
	r.Header.Add("Cookie", cookie)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resData, nil
}
