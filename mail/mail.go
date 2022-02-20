package mail

import (
	"github.com/go-marshaltemabu/go-unixtime"
)

type ReceiptAddress struct {
	Checked    string `json:"checked_address"`
	Normalized string `json:"normalized_address"`
}

type Mail struct {
	SenderAddress    string            `json:"sender_address"`
	ReceiptAddresses []ReceiptAddress  `json:"receipts"`
	ContentBody      string            `json:"body"`
	ReceiveAt        unixtime.UnixTime `json:"receive_at"`
}
