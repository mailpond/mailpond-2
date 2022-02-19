package mail

import (
	"github.com/go-marshaltemabu/go-unixtime"
)

type Mail struct {
	SenderAddress    string            `json:"sender"`
	ReceiptAddresses []string          `json:"receipts"`
	ContentBody      string            `json:"body"`
	ReceiveAt        unixtime.UnixTime `json:"receive_at"`
}
