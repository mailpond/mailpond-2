package server

import (
	"net/http"
	"time"

	unixtime "github.com/go-marshaltemabu/go-unixtime"
	utilhttphandlers "github.com/yinyin/go-util-http-handlers"

	"github.com/mailpond/mailpond-2"
)

type respGetMailsOfReceiptAddr struct {
	Mails   []*mailpond.Mail  `json:"mails"`
	QueryAt unixtime.UnixTime `json:"query_at"`
}

func (hnd *mailPondHTTPHandler) getMailsOfReceiptAddr(w http.ResponseWriter, r *http.Request, rcptAddr string) {
	mails, err := hnd.storageEngine.ListMails(r.Context(), rcptAddr)
	if nil != err {
		http.Error(w, "storage error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if mails == nil {
		mails = make([]*mailpond.Mail, 0)
	}
	utilhttphandlers.JSONResponseWithStatusOK(w, &respGetMailsOfReceiptAddr{
		Mails:   mails,
		QueryAt: unixtime.UnixTime(time.Now()),
	})
}
