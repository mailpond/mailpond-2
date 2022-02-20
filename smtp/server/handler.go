package server

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	smtp "github.com/emersion/go-smtp"
	unixtime "github.com/go-marshaltemabu/go-unixtime"
	emailnormalize "github.com/yinyin/go-email-address-normalize"

	"github.com/mailpond/mailpond-2"
)

const (
	maxRCPTProcessDuration = 3 * time.Second
	maxDATAProcessDuration = 10 * time.Second
)

var errAccountInvalidFormat = errors.New("account address is invalid")

type mailSession struct {
	storageEngine *mailpond.Storage

	senderAddress      string
	recipientAddresses []mailpond.ReceiptAddress
}

// Discard currently processed message.
func (s *mailSession) Reset() {
}

// Free all resources associated with session.
func (s *mailSession) Logout() (err error) {
	return
}

// Set return path for currently processed message.
func (s *mailSession) Mail(senderAddress string, opts smtp.MailOptions) (err error) {
	checkedAddr, _, err := emailnormalize.NormalizeEmailAddress(senderAddress, nil)
	if nil != err {
		log.Printf("WARN: cannot normalize sender address [%s]: %v", senderAddress, err)
		return
	}
	s.senderAddress = checkedAddr
	return
}

// Add recipient for currently processed message.
func (s *mailSession) Rcpt(recipientAddr string) (err error) {
	if len(recipientAddr) < 3 {
		err = errAccountInvalidFormat
		return
	}
	checkedAddr, normalizedAddr, err := emailnormalize.NormalizeEmailAddress(recipientAddr, nil)
	if nil != err {
		log.Printf("WARN: cannot normalize receipt address [%s]: %v", recipientAddr, err)
		return
	}
	rcptAddr := mailpond.ReceiptAddress{
		Checked:    checkedAddr,
		Normalized: normalizedAddr,
	}
	s.recipientAddresses = append(s.recipientAddresses, rcptAddr)
	return
}

// Set currently processed message contents and send it.
func (s *mailSession) Data(r io.Reader) (err error) {
	buf, err := io.ReadAll(r)
	if nil != err {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), maxDATAProcessDuration)
	defer cancel()
	m := mailpond.Mail{
		SenderAddress:    s.senderAddress,
		ReceiptAddresses: s.recipientAddresses,
		ContentBody:      string(buf),
		ReceiveAt:        unixtime.UnixTime(time.Now()),
	}
	err = s.storageEngine.AddMail(ctx, &m)
	return
}

type mailBackend struct {
	storageEngine *mailpond.Storage
}

func (b *mailBackend) createSession() (s smtp.Session, err error) {
	s = &mailSession{
		storageEngine: b.storageEngine,
	}
	return
}

// Authenticate a user. Return smtp.ErrAuthUnsupported if you don't want to
// support this.
func (b *mailBackend) Login(state *smtp.ConnectionState, username, password string) (s smtp.Session, err error) {
	err = smtp.ErrAuthUnsupported
	return
}

// Called if the client attempts to send mail without logging in first.
// Return smtp.ErrAuthRequired if you don't want to support this.
func (b *mailBackend) AnonymousLogin(state *smtp.ConnectionState) (s smtp.Session, err error) {
	return b.createSession()
}
