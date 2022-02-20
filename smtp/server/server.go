package server

import (
	"crypto/tls"
	"log"
	"sync"
	"time"

	smtp "github.com/emersion/go-smtp"

	"github.com/mailpond/mailpond-2"
)

func Serve(
	waitGroup *sync.WaitGroup,
	listenAddress,
	tlsCertFile, tlsKeyFile string,
	storageEngine *mailpond.Storage,
	domainAddress string,
	maxRecipients, maxMessageBytes int,
	readTimeout, writeTimeout time.Duration) (smtpServer *smtp.Server, err error) {
	backend := &mailBackend{
		storageEngine: storageEngine,
	}
	smtpServer = smtp.NewServer(backend)
	smtpServer.Addr = listenAddress
	if (tlsCertFile != "") && (tlsKeyFile != "") {
		var tlsCfg tls.Config
		tlsCfg.Certificates = make([]tls.Certificate, 1)
		if tlsCfg.Certificates[0], err = tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile); nil != err {
			return nil, err
		}
		smtpServer.TLSConfig = &tlsCfg
	}
	smtpServer.Domain = domainAddress
	smtpServer.AllowInsecureAuth = true
	smtpServer.MaxRecipients = maxRecipients
	smtpServer.MaxMessageBytes = maxMessageBytes
	smtpServer.ReadTimeout = readTimeout
	smtpServer.WriteTimeout = writeTimeout
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		var err error
		if (tlsCertFile != "") && (tlsKeyFile != "") {
			err = smtpServer.ListenAndServeTLS()
		} else {
			err = smtpServer.ListenAndServe()
		}
		log.Printf("INFO: stopped MailPond-2 SMTP service (err = %v)", err)
	}()
	return
}
