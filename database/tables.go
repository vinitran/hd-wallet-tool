package database

import (
	"HDWallet/wallet"
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"time"
)

type Token struct {
	bun.BaseModel `bun:"table:token,alias:tk"`
	Id            int       `bun:"id,pk,autoincrement" json:"id"`
	PrivateKey    string    `bun:"private_key,notnull" json:"private_key"`
	PublicKey     string    `bun:"public_key,notnull" json:"public_key"`
	Address       string    `bun:"address,notnull" json:"address"`
	Token         string    `bun:"token,notnull" json:"token"`
	Signature     string    `bun:"signature,notnull" json:"signature"`
	Nonce         string    `bun:"nonce,notnull" json:"nonce"`
	Session       string    `bun:"session,notnull" json:"session"`
	Time          time.Time `bun:"time,notnull" json:"time"`
}

func CreateTable(db *bun.DB) error {
	err := createRequestTokenTable(db)
	if err != nil {
		return err
	}

	return nil
}

func InsertTokenToTokenTable(db *bun.DB, data []wallet.Token) error {
	if data == nil {
		return nil
	}
	var tokens []Token
	for _, token := range data {
		timeParse, err := time.Parse(time.RFC3339, token.Signature.Time)
		if err != nil {
			return err
		}
		tokens = append(tokens, Token{
			PrivateKey: token.Signature.Wallet.PrivateKey,
			PublicKey:  token.Signature.Wallet.PublicKey,
			Address:    token.Signature.Wallet.Address,
			Token:      token.Token,
			Signature:  token.Signature.Signature,
			Nonce:      token.Signature.Nonce,
			Session:    token.Signature.Session,
			Time:       timeParse,
		})
	}
	_, err := db.NewInsert().
		Model(&tokens).
		Exec(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("database: inserted to db")
	return nil
}

func createRequestTokenTable(db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*Token)(nil)).
		IfNotExists().
		Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

var _ bun.AfterCreateTableHook = (*Token)(nil)

func (*Token) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	q := query.DB().NewCreateIndex().
		Model((*Token)(nil))
	err := addIndexTokenTable(q, "address_idx", "address")
	if err != nil {
		return err
	}

	err = addIndexTokenTable(q, "private_key_idx", "private_key")
	if err != nil {
		return err
	}
	_, err = q.IfNotExists().Exec(ctx)
	return err
}

func addIndexTokenTable(query *bun.CreateIndexQuery, index, column string) error {
	query = query.Index(index).Column(column)
	return nil
}
